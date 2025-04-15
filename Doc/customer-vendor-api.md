
### 1. Set Up Your Go Project
If you haven't done so already, initialize a new Go module:
```
go mod init adonai-api

```

### 2. Install Required Packages
Install the required packages:

```
go get github.com/dgrijalva/jwt-go
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/bson
go get go.mongodb.org/mongo-driver/mongo/options
go get golang.org/x/crypto/bcrypt
go get github.com/twilio/twilio-go
go get -u github.com/gorilla/mux
go get -u github.com/twilio/twilio-go

```
Sure, let's consolidate all the code and features into a single comprehensive comment. This will include setting up the backend (Golang), the frontend (React), and integrating the feeds page with the necessary APIs and UI components.

### Step-by-Step Guide

#### 1. Backend (Golang)

**Project Structure:**

```
adonai-api/
|-- config/
|   |-- config.go
|-- handlers/
|   |-- auth.go
|   |-- customer.go
|   |-- feed.go
|   |-- order.go
|   |-- store.go
|   |-- chat.go
|-- models/
|   |-- user.go
|   |-- customer.go
|   |-- store.go
|   |-- order.go
|   |-- feed.go
|   |-- chat.go
|-- middleware/
|   |-- auth.go
|-- main.go
```

##### `config/config.go`
Setup MongoDB connection.

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

func ConnectDB() {
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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

func InitEnv() {
    os.Setenv("JWT_SECRET_KEY", "your_jwt_secret_key")
}
```

##### `models/user.go`
User model definition.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Username string             `bson:"username" json:"username"`
    Password string             `bson:"password" json:"password"`
    Role     string             `bson:"role" json:"role"` // "customer" or "vendor"
    PhoneNumber  string         `bson:"phone_number" json:"phone_number"`
    OTP      string             `bson:"otp" json:"otp"`
    OTPExpiresAt int64          `bson:"otp_expires_at" json:"otp_expires_at"`
}
```

##### `models/customer.go`
Customer model definition.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Customer struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    FirstName string             `bson:"first_name" json:"first_name"`
    LastName  string             `bson:"last_name" json:"last_name"`
    StoreID   primitive.ObjectID `bson:"store_id,omitempty" json:"store_id,omitempty"`
}
```

##### `models/store.go`
Store model definition.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Store struct {
    ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name string             `bson:"name" json:"name"`
}
```

##### `models/order.go`
Order model definition.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID       primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    StoreID      primitive.ObjectID `bson:"store_id,omitempty" json:"store_id,omitempty"`
    Product      string             `bson:"product" json:"product"`
    Quantity     int                `bson:"quantity" json:"quantity"`
    OrderStatus  string             `bson:"order_status" json:"order_status"` // Pending, Delivered, Cancelled
    CreationDate int64              `bson:"creation_date" json:"creation_date"`
}
```

##### `models/feed.go`
Feed model definition.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Feed struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    Content   string             `bson:"content" json:"content"`
    CreatedAt int64              `bson:"created_at" json:"created_at"`
}
```

##### `models/chat.go`
Chat and Message models definition.

```go
package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type Message struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    FromUserID  primitive.ObjectID `bson:"from_user_id,omitempty" json:"from_user_id,omitempty"`
    ToAdmin     bool               `bson:"to_admin,omitempty" json:"to_admin,omitempty"`
    Content     string             `bson:"content" json:"content"`
    Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}

type Chat struct {
    UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    AdminID   primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
    Messages  []Message          `bson:"messages" json:"messages"`
}

type BroadcastMessage struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    AdminID   primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
    Content   string             `bson:"content" json:"content"`
    Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
```

##### `handlers/auth.go`
Auth handlers for signup and sign in with OTP.

``` go 
package middleware

import (
    "context"
    "adonai-api/handlers"
    "net/http"
    "strings"
    "time"

    "github.com/dgrijalva/jwt-go"
)

func JwtAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            if err == http.ErrNoCookie {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }

        tokenStr := cookie.Value
        claims := &handlers.Claims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return handlers.JwtKey, nil
        })

        if err != nil {
            if err == jwt.ErrSignatureInvalid {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }

        if !token.Valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userCtxValue := r.Context().Value("user")
            if userCtxValue == nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            userClaims := userCtxValue.(*handlers.Claims)
            if !strings.EqualFold(userClaims.Role, requiredRole) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}


```
##### `handlers/customer.go`
Customer-related handlers.

```go
package handlers

import (
    "context"
    "adonai-api/config"
    "adonai-api/models"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var customer models.Customer
    _ = json.NewDecoder(r.Body).Decode(&customer)
    collection := config.Client.Database("adonai-api").Collection("customers")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, _ := collection.InsertOne(ctx, customer)
    json.NewEncoder(w).Encode(result)
}

func GetCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    var customer models.Customer
    collection := config.Client.Database("adonai-api").Collection("customers")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    err := collection.FindOne(ctx, models.Customer{ID: id}).Decode(&customer)
    if err != nil {
        http.Error(w, "Customer not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(customer)
}

func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var customers []models.Customer
    collection := config.Client.Database("adonai-api").Collection("customers")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        var customer models.Customer
        cursor.Decode(&customer)
        customers = append(customers, customer)
    }
    json.NewEncoder(w).Encode(customers)
}

func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var customer models.Customer
    _ = json.NewDecoder(r.Body).Decode(&customer)
    collection := config.Client.Database("adonai-api").Collection("customers")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    filter := bson.M{"_id": id}
    update := bson.M{
        "$set": customer,
    }
    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        http.Error(w, "Customer not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(customer)
}

func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    collection := config.Client.Database("adonai-api").Collection("customers")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    _, err := collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        http.Error(w, "Customer not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("Customer deleted")
}
```

##### `handlers/feed.go`
Feed-related handlers.

```go
package handlers

import (
    "context"
    "adonai-api/config"
    "adonai-api/models"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateFeedHandler(w http.ResponseWriter, r *http.Request) {
    var feed models.Feed
    _ = json.NewDecoder(r.Body).Decode(&feed)
    feed.CreatedAt = time.Now().Unix()
    
    collection := config.Client.Database("adonai-api").Collection("feeds")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, err := collection.InsertOne(ctx, feed)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

func GetFeedsHandler(w http.ResponseWriter, r *http.Request) {
    userID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("user_id"))
    
    collection := config.Client.Database("adonai-api").Collection("feeds")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)
    
    var feeds []models.Feed
    for cursor.Next(ctx) {
        var feed models.Feed
        cursor.Decode(&feed)
        feeds = append(feeds, feed)
    }
    json.NewEncoder(w).Encode(feeds)
}
```
func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
    orderID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("order_id"))

    collection := config.Client.Database("adonai-api").Collection("orders")
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
    collection := config.Client.Database("adonai-api").Collection("orders")
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

package handlers

import (
    "context"
    "adonai-api/config"
    "adonai-api/models"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    var order models.Order
    _ = json.NewDecoder(r.Body).Decode(&order)
    order.CreationDate = time.Now().Unix()
    order.OrderStatus = "Pending"

    collection := config.Client.Database("adonai-api").Collection("orders")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    result, err := collection.InsertOne(ctx, order)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

func GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
    userID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("user_id"))

    collection := config.Client.Database("adonai-api").Collection("orders")
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

    collection



func BroadcastMessageHandler(w http.ResponseWriter, r *http.Request) {
    var msg models.BroadcastMessage
    _ = json.NewDecoder(r.Body).Decode(&msg)
    msg.Timestamp = time.Now()

    collection := config.Client.Database("adonai-api").Collection("broadcasts")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    _, err := collection.InsertOne(ctx, msg)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Optionally, distribute the broadcast message to all active users in real time.
    json.NewEncoder(w).Encode("Broadcast sent")
}
```

### Additional Middleware for Role-Based Access Control
#### `middleware/auth.go`
Update middleware to handle roles.

```go
package middleware

import (
    "context"
    "adonai-api/handlers"
    "net/http"
    "strings"
    "time"

    "github.com/dgrijalva/jwt-go"
)

func JwtAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            if err == http.ErrNoCookie {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }

        tokenStr := cookie.Value
        claims := &handlers.Claims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return handlers.JwtKey, nil
        })

        if err != nil {
            if err == jwt.ErrSignatureInvalid {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }

        if !token.Valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userCtxValue := r.Context().Value("user")
            if userCtxValue == nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            userClaims := userCtxValue.(*handlers.Claims)
            if !strings.EqualFold(userClaims.Role, requiredRole) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Update `main.go` to include new handlers and middleware
#### `main.go`

``` go 

package main

import (
    "adonai-api/config"
    "adonai-api/handlers"
    "adonai-api/middleware"
    "log"
    "net/http"
package main

import (
    "adonai-api/config"
    "adonai-api/handlers"
    "adonai-api/middleware"
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

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}

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

    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}


```
#### `chat.go`
``` go 
package handlers

import (
    "context"
    "adonai-api/config"
    "adonai-api/models"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
    var msg models.Message
    _ = json.NewDecoder(r.Body).Decode(&msg)
    msg.Timestamp = time.Now()

    collection := config.Client.Database("adonai-api").Collection("chats")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    // Check if a chat already exists between the user and admin
    chatCollection := config.Client.Database("adonai-api").Collection("chats")
    var chat models.Chat
    err := chatCollection.FindOne(ctx, bson.M{"user_id": msg.FromUserID}).Decode(&chat)

    if err != nil { // If no chat exists, create a new one
        chat = models.Chat{
            UserID:   msg.FromUserID,
            Messages: []models.Message{msg},
        }
        _, err = chatCollection.InsertOne(ctx, chat)
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
    } else { // If chat exists, update it with the new message
        chat.Messages = append(chat.Messages, msg)
        _, err = chatCollection.UpdateOne(ctx, bson.M{"_id": chat.ID}, bson.M{
            "$set": bson.M{"messages": chat.Messages},
        })
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
    }
    json.NewEncoder(w).Encode("Message sent")
}

func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
    userID, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("user_id"))

    collection := config.Client.Database("adonai-api").Collection("chats")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    var chat models.Chat
    err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&chat)
    if err != nil {
        http.Error(w, "Chat not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(chat.Messages)
}

func BroadcastMessageHandler(w http.ResponseWriter, r *http.Request) {
    var msg models.BroadcastMessage
    _ = json.NewDecoder(r.Body).Decode(&msg)
    msg.Timestamp = time.Now()

    collection := config.Client.Database("adonai-api").Collection("broadcasts")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

    _, err := collection.InsertOne(ctx, msg)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Optionally, distribute the broadcast message to all active users in real time.
    json.NewEncoder(w).Encode("Broadcast sent")
}

```


### Frontend (React) App

To create the frontend, we will use React.js to handle user interactions and communicate with the backend API.

#### Create a new React app

Install create-react-app if you haven’t already:

```sh
npx create-react-app customer-vendor-app
cd customer-vendor-app
```

Install necessary dependencies:

```sh
npm install axios react-router-dom socket.io-client
```

#### Structure
```
src/
|-- api/
|   |-- index.js
|-- components/
|   |-- AdminDashboard.js
|   |-- Chat.js
|   |-- CustomerDashboard.js
|   |-- Login.js
|   |-- Signup.js
|   |-- OrderForm.js
|-- App.js
|-- index.js
```

#### `src/api/index.js`

Set up API requests using Axios.

```js
import axios from "axios";

const API_URL = "http://localhost:8080";

export const signup = (data) => axios.post(`${API_URL}/signup`, data);
export const requestOTP = (data) => axios.post(`${API_URL}/request-otp`, data);
export const verifyOTP = (data) => axios.post(`${API_URL}/verify-otp`, data);

export const createOrder = (data) => axios.post(`${API_URL}/order`, data);
export const getUserOrders = (userId) => axios.get(`${API_URL}/orders`, { params: { user_id: userId } });
export const cancelOrder = (orderId) => axios.put(`${API_URL}/cancel-order`, null, { params: { order_id: orderId } });

export const getAllOrders = () => axios.get(`${API_URL}/all-orders`);

export const sendMessage = (data) => axios.post(`${API_URL}/send-message`, data);
export const getChatHistory = (userId) => axios.get(`${API_URL}/chat-history`, { params: { user_id: userId } });
export const broadcastMessage = (data) => axios.post(`${API_URL}/broadcast`, data);
```

#### `src/components/Signup.js`

Component for user signup.

```js
import React, { useState } from "react";
import { signup } from "../api";

export default function Signup() {
  const [formData, setFormData] = useState({
    username: "",
    password: "",
    phone_number: "",
    role: "customer", // default role
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData, [name]: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await signup(formData);
      alert("Signup successful, please verify OTP sent to your phone number.");
    } catch (error) {
      console.error(error);
      alert("Signup failed. Please try again.");
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label>Username:</label>
        <input type="text" name="username" onChange={handleChange} value={formData.username} required />
      </div>
      <div>
        <label>Password:</label>
        <input type="password" name="password" onChange={handleChange} value={formData.password} required />
      </div>
      <div>
        <label>Phone Number:</label>
        <input type="text" name="phone_number" onChange={handleChange} value={formData.phone_number} required />
      </div>
      <div>
        <label>Role:</label>
        <select name="role" onChange={handleChange} value={formData.role}>
          <option value="customer">Customer</option>
          <option value="vendor">Vendor</option>
        </select>
      </div>
      <button type="submit">Sign Up</button>
    </form>
  );
}
```


Certainly! Below is the continuation and completion of the React components:

#### `src/components/Login.js`

```js
import React, { useState } from "react";
import { requestOTP, verifyOTP } from "../api";

export default function Login({ setRole }) {
  const [phoneNumber, setPhoneNumber] = useState("");
  const [otp, setOtp] = useState("");
  const [otpRequested, setOtpRequested] = useState(false);

  const handleRequestOTP = async () => {
    try {
      await requestOTP({ phone_number: phoneNumber });
      setOtpRequested(true);
      alert("OTP sent to your phone number.");
    } catch (error) {
      console.error(error);
      alert("Failed to request OTP. Please try again.");
    }
  };

  const handleVerifyOTP = async () => {
    try {
      // Assuming the server returns user data upon OTP verification
      const response = await verifyOTP({ phone_number: phoneNumber, otp });
      setRole(response.data.role);
      alert("Login successful.");
    } catch (error) {
      console.error(error);
      alert("Invalid OTP. Please try again.");
    }
  };

  return (
    <div>
      {!otpRequested ? (
        <div>
          <label>Phone Number:</label>
          <input
            type="text"
            value={phoneNumber}
            onChange={(e) => setPhoneNumber(e.target.value)}
            required
          />
          <button onClick={handleRequestOTP}>Request OTP</button>
        </div>
      ) : (
        <div>
          <label>OTP:</label>
          <input
            type="text"
            value={otp}
            onChange={(e) => setOtp(e.target.value)}
            required
          />
          <button onClick={handleVerifyOTP}>Verify OTP</button>
        </div>
      )}
    </div>
  );
}
```

#### `src/components/CustomerDashboard.js`

Component for the customer dashboard.

```js
import React, { useEffect, useState } from "react";
import { createOrder, getUserOrders, cancelOrder } from "../api";

export default function CustomerDashboard({ userId }) {
  const [orders, setOrders] = useState([]);
  const [newOrder, setNewOrder] = useState({ product: "", quantity: "" });

  useEffect(() => {
    async function fetchOrders() {
      try {
        const response = await getUserOrders(userId);
        setOrders(response.data);
      } catch (error) {
        console.error(error);
      }
    }
    fetchOrders();
  }, [userId]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setNewOrder((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };

  const handleCreateOrder = async (e) => {
    e.preventDefault();
    try {
      await createOrder({ ...newOrder, user_id: userId });
      alert("Order created successfully.");
      setNewOrder({ product: "", quantity: "" });
      await fetchOrders(); // Fetch the updated orders
    } catch (error) {
      console.error(error);
      alert("Failed to create order. Please try again.");
    }
  };

  const handleCancelOrder = async (orderId) => {
    try {
      await cancelOrder(orderId);
      alert("Order cancelled successfully.");
      await fetchOrders(); // Fetch the updated orders
    } catch (error) {
      console.error(error);
      alert("Failed to cancel order. Please try again.");
    }
  };

  return (
    <div>
      <h2>Customer Dashboard</h2>
      <h3>Create New Order</h3>
      <form onSubmit={handleCreateOrder}>
        <div>
          <label>Product:</label>
          <input
            type="text"
            name="product"
            onChange={handleChange}
            value={newOrder.product}
            required
          />
        </div>
        <div>
          <label>Quantity:</label>
          <input
            type="number"
            name="quantity"
            onChange={handleChange}
            value={newOrder.quantity}
            required
          />
        </div>
        <button type="submit">Create Order</button>
      </form>
      <h3>Previous Orders</h3>
      <ul>
        {orders.map((order) => (
          <li key={order.id}>
            {order.product} - {order.quantity} - {order.order_status}
            {order.order_status === "Pending" && (
              <button onClick={() => handleCancelOrder(order.id)}>Cancel</button>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

#### `src/components/AdminDashboard.js`

Component for the admin dashboard.

```js
import React, { useEffect, useState } from "react";
import { getAllOrders, broadcastMessage } from "../api";

export default function AdminDashboard() {
  const [orders, setOrders] = useState([]);
  const [broadcastContent, setBroadcastContent] = useState("");

  useEffect(() => {
    async function fetchAllOrders() {
      try {
        const response = await getAllOrders();
        setOrders(response.data);
      } catch (error) {
        console.error(error);
      }
    }
    fetchAllOrders();
  }, []);

  const handleBroadcastChange = (e) => {
    setBroadcastContent(e.target.value);
  };

  const handleBroadcast = async (e) => {
    e.preventDefault();
    try {
      await broadcastMessage({ content: broadcastContent });
      alert("Broadcast message sent successfully.");
      setBroadcastContent("");
    } catch (error) {
      console.error(error);
      alert("Failed to send broadcast message. Please try again.");
    }
  };

  return (
    <div>
      <h2>Admin Dashboard</h2>
      <h3>All Orders</h3>
      <ul>
        {orders.map((order) => (
          <li key={order.id}>
            {order.product} - {order.quantity} - {order.order_status} - {order.user_id}
          </li>
        ))}
      </ul>
      <h3>Broadcast Message</h3>
      <form onSubmit={handleBroadcast}>
        <textarea
          value={broadcastContent}
          onChange={handleBroadcastChange}
          required
        />
        <button type="submit">Send Broadcast</button>
      </form>
    </div>
  );
}
```

#### `src/components/Chat.js`

Component for live chat.

```js
import React, { useEffect, useState } from "react";
import { sendMessage, getChatHistory } from "../api";

export default function Chat({ userId }) {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");

  useEffect(() => {
    async function fetchChatHistory() {
      try {
        const response = await getChatHistory(userId);
        setMessages(response.data);
      } catch (error) {
        console.error(error);
      }
    }
    fetchChatHistory();
  }, [userId]);

  const handleMessageChange = (e) => {
    setNewMessage(e.target.value);
  };

  const handleMessageSend = async (e) => {
    e.preventDefault();
    try {
      await sendMessage({ from_user_id: userId, content: newMessage, to_admin: true });
      setMessages([...messages, { content: newMessage, to_admin: true }]);
      setNewMessage("");
    } catch (error) {
      console.error(error);
      alert("Failed to send message. Please try again.");
    }
  };

  return (
    <div>
      <h2>Chat</h2>
      <div className="chat-window">
        {messages.map((msg, index) => (
          <div key={index} className={msg.to_admin ? "message sent" : "message received"}>
            {msg.content}
          </div>
        ))}
      </div>
      <form onSubmit={handleMessageSend}>
        <input
          type="text"
          value={newMessage}
          onChange={handleMessageChange}
          required
        />
        <button type="submit">Send</button>
      </form>
    </div>
  );
}
```

#### `src/App.js`

Main App component to handle routing and role-based rendering.

```js
import React, { useState } from "react";
import { BrowserRouter as Router, Route, Switch, Redirect } from "react-router-dom";
import Signup from "./components/Signup";
import Login from "./components/Login";
import CustomerDashboard from "./components/CustomerDashboard";
import AdminDashboard from "./components/AdminDashboard";
import Chat from "./components/Chat";

function App() {
  const [role, setRole] = useState("");
  const [userId, setUserId] = useState("");

  if (!role) {
    return (
      <Router>
        <Switch>
          <Route path="/signup">
            <Signup />
          </Route>
          <Route path="/login">
            <Login setRole={setRole} setUserId={setUserId} />
          </Route>
          <Redirect to="/login" />
        </Switch>
      </Router>
    );
  }

  return (
    <Router>
      <div className="App">
        {role === "customer" && (
          <>
            <CustomerDashboard userId={userId} />
            <Chat userId={userId} />
          </>
        )}
        {role === "vendor" && <AdminDashboard />}
      </div>
    </Router>
  );
}

export default App;
```

#### `src/index.js`

Entry point to the React application.

```js
import React from "react";
import ReactDOM from "react-dom";
import "./index.css";
import App from "./App";

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById("root")
);
```

### Explanation:
1. **Signup Component** allows users to create new accounts.
2. **Login Component** handles phone number-based OTP authentication and sets the user role.
3. **CustomerDashboard Component** allows customers to manage orders.
4. **AdminDashboard Component** allows vendors to view all orders and send broadcast messages.
5. **Chat Component** enables live chat between customers and admins.
6. **App Component** sets up routing and renders components based on user role and authentication status.

This setup provides a basic implementation of the requested features. You can expand it further based on your specific requirements, such as integrating with WebSocket for real-time chat updates, adding more comprehensive error handling, and improving the UI/UX.