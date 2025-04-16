To support storing and sharing images in the Order model, and ensuring images are stored in an optimized way, we'll make use of cloud storage (e.g., AWS S3) to store images and store only the image URLs in the MongoDB database. This approach will help optimize storage and delivery.

### Step-by-Step Guide:

1. **Update the Order model to include an image property.**
2. **Implement image upload handling.**
3. **Store the image in cloud storage (AWS S3).**
4. **Update handlers to handle image uploads.**

### 1. Update the Order Model

Add an `ImageURL` field to the Order model.

**models/order.go:**
```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    StoreID      primitive.ObjectID `bson:"store_id,omitempty" json:"store_id,omitempty"`
    Product      string             `bson:"product" json:"product"`
    Quantity     int                `bson:"quantity" json:"quantity"`
    OrderStatus  string             `bson:"order_status" json:"order_status"`
    CreationDate int64              `bson:"creation_date" json:"creation_date"`
    ImageURL     string             `bson:"image_url" json:"image_url"`
}
```

### 2. Set Up AWS S3

1. **Create an S3 Bucket:**
   - Sign in to the AWS Management Console.
   - Navigate to S3 and create a new bucket.
   - Note the bucket name and region.

2. **Create an IAM User:**
   - Go to the IAM console and create a new user with programmatic access. Attach policies for S3 access.

3. **Store AWS Credentials:**
   - Store the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY as environment variables.

### 3. Install Necessary Packages

Install the AWS SDK for Go.

```sh
go get github.com/aws/aws-sdk-go/aws
go get github.com/aws/aws-sdk-go/aws/session
go get github.com/aws/aws-sdk-go/service/s3
```

### 4. Implement Image Upload Handling

**handlers/order.go:**
```go
package handlers

import (
    "context"
    "customer_vendor_api/config"
    "customer_vendor_api/models"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Initialize S3 Session
var (
    S3Region    = "your-region"
    S3Bucket    = "your-bucket-name"
    S3AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
    S3SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func uploadImageToS3(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(S3Region),
        Credentials: credentials.NewStaticCredentials(
            S3AccessKey, S3SecretKey, ""),
    })
    if err != nil {
        return "", err
    }

    defer file.Close()

    buffer := make([]byte, fileHeader.Size)
    file.Read(buffer)

    filePath := fmt.Sprintf("orders/%d-%s", time.Now().Unix(), fileHeader.Filename)
    _, err = s3.New(sess).PutObject(&s3.PutObjectInput{
        Bucket: aws.String(S3Bucket),
        Key:    aws.String(filePath),
        Body:   aws.ReadSeekCloser(bytes.NewReader(buffer)),
        ACL:    aws.String("public-read"),
    })
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3Bucket, S3Region, filePath), nil
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // 10 MB

    file, fileHeader, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    
    imageURL, err := uploadImageToS3(file, fileHeader)
    if err != nil {
        log.Printf("Failed to upload image: %v", err)
        http.Error(w, "Failed to upload image", http.StatusInternalServerError)
        return
    }

    var order models.Order
    _ = json.NewDecoder(r.FormFile("payload")).Decode(&order)
    order.CreationDate = time.Now().Unix()
    order.OrderStatus = "Pending"
    order.ImageURL = imageURL

    collection := config.Client.Database("customer_vendor_api").Collection("orders")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    result, err := collection.InsertOne(ctx, order)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

// Other handlers remain similar

func GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
    userID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("user_id"))

    collection := config.Client.Database("customer_vendor_api").Collection("orders")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var orders []models.Order
    for cursor.Next(ctx) {
        var order models.Order
        cursor.Decode(&order)
        orders = append(orders, order)
    }
    json.NewEncoder(w).Encode(orders)
}

func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
    orderID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("order_id"))

    collection := config.Client.Database("customer_vendor_api").Collection("orders")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

    _, err := collection.UpdateOne(ctx, bson.M{"_id": orderID}, bson.M{
        "$set": bson.M{"order_status": "Cancelled"},
    })
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode("Order cancelled")
}

func GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
    collection := config.Client.Database("customer_vendor_api").Collection("orders")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var orders []models.Order
    for cursor.Next(ctx) {
        var order models.Order
        cursor.Decode(&order)
        orders = append(orders, order)
    }
    json.NewEncoder(w).Encode(orders)
}
```

**Explanation:**

1. **S3 Configuration:**
   - `S3Region`, `S3Bucket`, `S3AccessKey`, and `S3SecretKey`: Store S3 configuration details and credentials.

2. **uploadImageToS3**:
   - Creates an S3 session using access keys.
   - Reads the file and uploads it to a specific path in the S3 bucket.
   - Returns the public URL of the uploaded image.

3. **CreateOrderHandler**:
   - Parses multipart form data to handle file uploads.
   - Uploads the image to S3 using the `uploadImageToS3` function.
   - Stores the order data along with the image URL into MongoDB.

4. **Other Handlers**: Functions like `GetUserOrdersHandler`, `CancelOrderHandler`, and `GetAllOrdersHandler` remain similar to previous examples but now recognize and handle the `ImageURL` field in the `Order` struct.

### Update frontend to handle image uploads correctly

**HTML Form Example:**
```html
<form id="orderForm" method="POST" enctype="multipart/form-data">
  <input type="file" name="image" />
  <input type="text" name="payload" value='{"product": "test product", "quantity": 1}' />
  <button type="submit">Submit Order</button>
</form>
```

### Conclusion

With these changes, images associated with orders are now stored in an optimized way using AWS S3, and only the URLs pointing to these images are stored in MongoDB. This strategy ensures efficient image storage and retrieval while keeping our database lightweight. By separating storage concerns and leveraging cloud storage, we improve the scalability and performance of our application.



## Logic to save and render asset images such as logos and other PNG files. To optimize logos for faster loading, we'll use image optimization techniques such as resizing the image upon upload and using compressed image formats. We'll use AWS S3 for storing the images and MongoDB for storing image metadata.

### Step-by-Step Guide:

1. **Update the Model to Include Asset Images**
2. **Implement Image Upload and Optimization Logic**
3. **Create Handlers for Uploading and Retrieving Images**
4. **Setup AWS S3 for Storage**
5. **Optimize Image Loading**

### 1. Update the Model to Include Asset Images

Add an `Asset` model to store image metadata.

**models/asset.go:**
```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Asset struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name      string             `bson:"name" json:"name"`
    URL       string             `bson:"url" json:"url"`
    Type      string             `bson:"type" json:"type"` // e.g. logo, banner
    CreatedAt int64              `bson:"created_at" json:"created_at"`
}
```

### 2. Implement Image Upload and Optimization Logic

**handlers/asset.go:**
```go
package handlers

import (
    "bytes"
    "context"
    "customer_vendor_api/config"
    "customer_vendor_api/models"
    "encoding/json"
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/nfnt/resize"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
    S3Region    = "your-region"
    S3Bucket    = "your-bucket-name"
    S3AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
    S3SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

// uploadImageToS3 optimized and upload
func uploadImageToS3(file multipart.File, fileHeader *multipart.FileHeader, assetType string) (string, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(S3Region),
        Credentials: credentials.NewStaticCredentials(
            S3AccessKey, S3SecretKey, ""),
    })
    if err != nil {
        return "", err
    }

    defer file.Close()

    var buffer bytes.Buffer
    var contentType string

    if strings.HasSuffix(fileHeader.Filename, ".png") {
        img, err := png.Decode(file)
        if err != nil {
            return "", err
        }

        // Resize the image to width 400 preserving aspect ratio
        img = resize.Resize(400, 0, img, resize.Lanczos3)

        contentType = "image/png"
        err = png.Encode(&buffer, img)

    } else if strings.HasSuffix(fileHeader.Filename, ".jpg") || strings.HasSuffix(fileHeader.Filename, ".jpeg") {
        img, err := jpeg.Decode(file)
        if err != nil {
            return "", err
        }

        // Resize the image to width 400 preserving aspect ratio
        img = resize.Resize(400, 0, img, resize.Lanczos3)

        contentType = "image/jpeg"
        err = jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 80})

    } else {
        return "", fmt.Errorf("unsupported file format")
    }

    if err != nil {
        return "", err
    }

    filePath := fmt.Sprintf("assets/%s/%d-%s", assetType, time.Now().Unix(), fileHeader.Filename)
    _, err = s3.New(sess).PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(S3Bucket),
        Key:         aws.String(filePath),
        Body:        aws.ReadSeekCloser(bytes.NewReader(buffer.Bytes())),
        ContentType: aws.String(contentType),
        ACL:         aws.String("public-read"),
    })
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3Bucket, S3Region, filePath), nil
}

// UploadAssetHandler handles uploading of asset images
func UploadAssetHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // 10 MB

    file, fileHeader, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }

    assetType := r.FormValue("type") // type such as "logo" or "banner"
    if assetType == "" {
        http.Error(w, "Asset type is required", http.StatusBadRequest)
        return
    }

    imageURL, err := uploadImageToS3(file, fileHeader, assetType)
    if err != nil {
        log.Printf("Failed to upload image: %v", err)
        http.Error(w, "Failed to upload image", http.StatusInternalServerError)
        return
    }

    asset := models.Asset{
        Name:      fileHeader.Filename,
        URL:       imageURL,
        Type:      assetType,
        CreatedAt: time.Now().Unix(),
    }

    collection := config.Client.Database("customer_vendor_api").Collection("assets")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    result, err := collection.InsertOne(ctx, asset)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

// GetAssetHandler retrieves a specific asset
func GetAssetHandler(w http.ResponseWriter, r *http.Request) {
    assetID := r.URL.Query().Get("id")
    if assetID == "" {
        http.Error(w, "Asset ID is required", http.StatusBadRequest)
        return
    }

    id, err := primitive.ObjectIDFromHex(assetID)
    if err != nil {
        http.Error(w, "Invalid Asset ID", http.StatusBadRequest)
        return
    }

    collection := config.Client.Database("customer_vendor_api").Collection("assets")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

    var asset models.Asset
    err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&asset)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "Asset not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(asset)
}

// ListAssetsHandler lists all assets
func ListAssetsHandler(w http.ResponseWriter, r *http.Request) {
    collection := config.Client.Database("customer_vendor_api").Collection("assets")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var assets []models.Asset
    for cursor.Next(ctx) {
        var asset models.Asset
        cursor.Decode(&asset)
        assets = append(assets, asset)
    }
    json.NewEncoder(w).Encode(assets)
}
```

**Explanation:**

1. **Models** (`models/order.go`, `models/asset.go`):
   - Updated the `Order` model to include `ImageURL`.
   - Added a new `Asset` model to manage different types of asset images.

2. **AWS S3 Configuration**:
   - Setting up AWS S3 configuration to store and retrieve images.

3. **Image Upload and Optimization**:
   - Added `uploadImageToS3` function to handle file uploads to S3 with image resizing and optimization.
   - Supported image formats: PNG and JPEG.
   - Used `github.com/nfnt/resize` for resizing images.

4. **Handlers** (`handlers/asset.go`):
   - Added `UploadAssetHandler` to handle image uploads.
   - Added `GetAssetHandler` to retrieve a specific asset image by its ID.
   - Added `ListAssetsHandler` to list all asset images.

### 4. Update main.go to Include Asset Routes

**main.go**
```go
package main

import (
    "customer_vendor_api/config"
    "customer_vendor_api/handlers"
    "customer_vendor_api/middleware"
    "log"
    "net/http"

    "github.com/gorilla/mux"
)

func main() {
    config.InitEnv()
    config.ConnectDB()

    r := mux.NewRouter()

    // Authentication routes
    r.HandleFunc("/signup", handlers.SignUpHandler).Methods("POST")
    r.HandleFunc("/request-otp", handlers.RequestOTPHandler).Methods("POST")
    r.HandleFunc("/verify-otp", handlers.VerifyOTPHandler).Methods("POST")

    // Customer routes
    r.Handle("/customers", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetCustomersHandler))).Methods("GET")
    r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateCustomerHandler))).Methods("POST")
    r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetCustomerHandler))).Methods("GET")
    r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.UpdateCustomerHandler))).Methods("PUT")
    r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.DeleteCustomerHandler))).Methods("DELETE")

    // Store routes
    r.Handle("/stores", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetStoresHandler))).Methods("GET")
    r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateStoreHandler))).Methods("POST")
    r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetStoreHandler))).Methods("GET")
    r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.UpdateStoreHandler))).Methods("PUT")
    r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.DeleteStoreHandler))).Methods("DELETE")

    // Order routes
    r.Handle("/orders", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetUserOrdersHandler))).Methods("GET")
    r.Handle("/order", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateOrderHandler))).Methods("POST")
    r.Handle("/cancel-order", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CancelOrderHandler))).Methods("PUT")
    r.Handle("/all-orders", middleware.JwtAuthMiddleware(middleware.RoleMiddleware("vendor")(http.HandlerFunc(handlers.GetAllOrdersHandler)))).Methods("GET")

    // Chat routes
    r.Handle("/send-message", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.SendMessageHandler))).Methods("POST")
    r.Handle("/chat-history", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetChatHistoryHandler))).Methods("GET")
    r.Handle("/broadcast", middleware.JwtAuthMiddleware(middleware.RoleMiddleware("vendor")(http.HandlerFunc(handlers.BroadcastMessageHandler)))).Methods("POST")

    // Feed routes
    r.Handle("/feeds", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetFeedsHandler))).Methods("GET")
    r.Handle("/feed", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateFeedHandler))).Methods("POST")

    // Email notification route
    r.Handle("/send-email", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.SendEmailHandler))).Methods("POST")

    // Asset routes - New routes for handling assets
    r.Handle("/upload-asset", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.UploadAssetHandler))).Methods("POST")
    r.Handle("/asset", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetAssetHandler))).Methods("GET")
    r.Handle("/assets", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.ListAssetsHandler))).Methods("GET")

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}
```

### Conclusion

With these changes, the server can now handle the upload, storage, and retrieval of optimized asset images, such as logos and other PNGs. By using AWS S3 for storage and employing image optimization techniques, we ensure that the images are stored and delivered efficiently. This makes the application more scalable and improves the user experience with faster image loading times.