Walk you through the entire backend logic of our application, explaining the Go (Golang) syntax and code structure as if you're new to programming. We'll focus on the core tasks: handling user requests, connecting to a database, and sending emails. 

### Overview

Our backend application serves as an intermediary between the user and the database. It listens for requests, processes the required actions (like registering a user or sending an email), and returns the necessary responses.

### Project Structure

Here’s a simplified view of our project's structure:

```
customer_vendor_api/
├── config/
│   └── config.go
├── handlers/
│   ├── auth.go
│   ├── customer.go
│   ├── feed.go
│   ├── order.go
│   ├── store.go
│   ├── chat.go
│   └── email.go
├── middleware/
│   └── auth.go
├── models/
│   ├── user.go
│   ├── customer.go
│   ├── store.go
│   ├── order.go
│   ├── feed.go
│   └── chat.go
├── main.go
├── go.mod
```

### Main Concepts

1. **Handlers**: Functions that manage specific tasks when users request them.
2. **Routes**: Paths associated with handlers, defining what each path should do.
3. **Database Connection**: Connects our app to a database to store and fetch information.
4. **Middleware**: Functions that work before the main handlers to perform preliminary checks (e.g., authentication).
5. **Environment Variables**: Secure storage for sensitive data like API keys.

### Detailed Explanation

#### `main.go`

This is the entry point of our application.

1. **Importing Packages**: We import the necessary packages providing various functionalities like routing, logging, HTTP, and configuration.

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
```

2. **Main Function**: Like the engine starting a car, this function begins our application's execution.

- **Initialize Environment and Database**: We set up environment variables and connect to MongoDB (the database).
- **Setting Up Routes with Handlers**: We define different paths (routes) and associate them with specific functions (handlers) for actions like user signup, order creation, etc.

```go
func main() {
    config.InitEnv()
    config.ConnectDB()

    r := mux.NewRouter()

    // Authentication routes
    r.HandleFunc("/signup", handlers.SignUpHandler).Methods("POST")
    r.HandleFunc("/request-otp", handlers.RequestOTPHandler).Methods("POST")
    r.HandleFunc("/verify-otp", handlers.VerifyOTPHandler).Methods("POST")

    // Customer routes - Routes are paths that users can access.
    r.Handle("/customers", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetCustomersHandler))).Methods("GET")
    // more routes here...

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}
```

### `config/config.go`

This file sets up the database connection.

1. **Connecting to MongoDB**:

We specify the MongoDB URI (location) and connect to the database. The `Ping` function checks if the connection to the database is successful.

```go
package config

import (
    "context"
    "fmt"
    "log"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() {
    uri := "mongodb://localhost:27017"
    clientOptions := options.Client().ApplyURI(uri).SetServerSelectionTimeout(10 * time.Second)

    var err error
    Client, err = mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = Client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")
}
```

### `handlers/auth.go`

Handles authentication-related tasks like user signup, OTP requests, and verification.

1. **Declaring Variables and Structs**:

```go
package handlers

import (
    "context"
    "customer_vendor_api/config"
    "customer_vendor_api/models"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "os"
    "time"
    "github.com/dgrijalva/jwt-go"
    "github.com/twilio/twilio-go"
    openapi "github.com/twilio/twilio-go/rest/api/v2010"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))  // Secure storage for our key

type Credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type OTPRequest struct {
    PhoneNumber string `json:"phone_number"`
    OTP         string `json:"otp"`
}

type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.StandardClaims
}

func generateOTP() string {
    rand.Seed(time.Now().UnixNano())
    return fmt.Sprintf("%06d", rand.Intn(1000000))
}
```

2. **Handler Functions**:

#### SignUpHandler

- **Receiving User Data**: Uses `json.NewDecoder(r.Body).Decode(&user)` to read the user's data sent in the request.
- **Password Hashing**: Converts the password to a hashed version for security using `bcrypt`.
- **Storing Data**: Saves the new user's data to the database.

```go
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
    var user models.User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    user.Password = string(hashedPassword)

    collection := config.Client.Database("customer_vendor_api").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    result, err := collection.InsertOne(ctx, user)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}
```

#### `handlers/email.go`

Handles sending email notifications.

1. **Sending an Email**:

- **Read Request**: Understand user request via JSON.
- **Create and Send Email**: Using SendGrid, create the email's content and send it.

```go
package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

var sendGridAPIKey = os.Getenv("SENDGRID_API_KEY")

type EmailRequest struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func SendEmailHandler(w http.ResponseWriter, r *http.Request) {
    var emailReq EmailRequest
    err := json.NewDecoder(r.Body).Decode(&emailReq)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    from := mail.NewEmail("Your App Name", "you@example.com")
    to := mail.NewEmail("", emailReq.To)
    plainTextContent := emailReq.Body
    htmlContent := "<p>" + emailReq.Body + "</p>"
    message := mail.NewSingleEmail(from, emailReq.Subject, to, plainTextContent, htmlContent)

    client := sendgrid.NewSendClient(sendGridAPIKey)
    response, err := client.Send(message)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to send email: %s", err.Error()), http.StatusInternalServerError)
        return
    }

    fmt.Printf("Email sent successfully with status code %d\n", response.StatusCode)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"})
}
```

### Conclusion

In a nutshell:

- **Handlers**: Respond to user requests. For instance, when a user signs up, `SignUpHandler` manages the request.
- **Routes**: Connect specific paths (like `/signup`) to their handlers.
- **Database Connection**: Ensures our application can store and retrieve information.
- **Middleware**: Adds additional checks before handlers run (like verifying a user is authenticated).
- **Environment Variables**: Keep sensitive information secure.

Go is efficient and built to handle tasks like web development seamlessly. Each part of our application—from the handlers managing HTTP requests to the middleware ensuring additional checks—works together to create a secure, efficient backend. This backend processes user input and interacts with the database, making our entire application functional.

This explanation should give you a comprehensive understanding of how the backend logic and Go syntax work together, even if you are new to programming.





## Local MongoDB setup with MongoDB Atlas (a cloud-hosted MongoDB service) and MongoDB Compass (a GUI for MongoDB).

### Overview

1. **Create a MongoDB Atlas Cluster**
2. **Connect to MongoDB Atlas via MongoDB Compass**
3. **Update Go application to use MongoDB Atlas**

### Step 1: Create a MongoDB Atlas Cluster

1. **Sign up or Log in to MongoDB Atlas**
   - Visit [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) and sign up or log in.

2. **Create a Cluster**
   - Follow the instructions to create a new free-tier cluster.

3. **Add a Database User**
   - Under the Security tab, add a new database user with a username and password.

4. **Whitelist Your IP Address**
   - Add your IP address to the whitelist under the Network Access tab.

5. **Get Connection String**
   - In the Cluster view, click "Connect" and choose "Connect your application". Copy the connection string.

### Step 2: Connect to MongoDB Atlas via MongoDB Compass

1. **Install MongoDB Compass**
   - Download and install MongoDB Compass from: [MongoDB Compass Download](https://www.mongodb.com/products/compass)

2. **Connect MongoDB Compass to Atlas**
   - Open MongoDB Compass and paste the connection string you copied from Atlas. Replace `<password>` with the password of the database user you created earlier. Click "Connect".

### Step 3: Update Go Application to Use MongoDB Atlas

1. **Update Configuration to Use MongoDB Atlas Connection String**

**config/config.go:**
```go
package config

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func InitEnv() {
    // Load environment variables from a .env file if necessary (optional)
    // _ = godotenv.Load()

    // Ensure that the necessary environment variables are set
    os.Setenv("MONGODB_URI", "your_mongo_db_atlas_connection_string")
}

func ConnectDB() {
    uri := os.Getenv("MONGODB_URI")
    
    clientOptions := options.Client().ApplyURI(uri).SetServerSelectionTimeout(10 * time.Second)

    var err error
    Client, err = mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = Client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")
}
```

Replace `your_mongo_db_atlas_connection_string` with the connection string you got from MongoDB Atlas, but replace `<password>` with the actual password of the database user.

2. **Ensure Go Application Imports and Runs Correctly**

**main.go:**
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

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}
```

### Conclusion

By following these steps, you have switched from a local MongoDB setup to using MongoDB Atlas, a managed cloud database service. You can manage and visualize your database using MongoDB Compass. This setup is more scalable and reliable, as MongoDB Atlas handles the underlying infrastructure and scaling for you. Remember to handle your environment variables securely, and avoid hardcoding sensitive information in your source code. Use a `.env` file or secret management systems provided by cloud platforms for better security.