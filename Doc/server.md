Sure, let's create a simple CRUD (Create, Read, Update, Delete) operation example using Golang and MongoDB. We'll use the `go.mongodb.org/mongo-driver` package for MongoDB operations. Below is a step-by-step guide to set up and implement the CRUD operations.

### 1. Set Up Your Go Project

Initialize a new Go module:

```sh
go mod init crudexample
```

### 2. Install MongoDB Go Driver

Install the MongoDB Go driver:

```sh
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
go get go.mongodb.org/mongo-driver/mongo/bson
```

### 3. Implement CRUD Operations

Create the following structure for the project:

```
crudexample/
|-- main.go
|-- models/
    |-- user.go
|-- handlers/
    |-- user.go
```

#### `models/user.go`

This file will contain the User model.

```go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    FirstName string             `bson:"firstName" json:"firstName"`
    LastName  string             `bson:"lastName" json:"lastName"`
    Email     string             `bson:"email" json:"email"`
}
```

#### `handlers/user.go`

This file will contain the handlers for the User CRUD operations.

```go
package handlers

import (
    "context"
    "crudexample/models"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func ConnectDB() {
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    client, err = mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var user models.User
    _ = json.NewDecoder(r.Body).Decode(&user)
    collection := client.Database("crudexample").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, _ := collection.InsertOne(ctx, user)
    json.NewEncoder(w).Encode(result)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    var user models.User
    collection := client.Database("crudexample").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    err := collection.FindOne(ctx, models.User{ID: id}).Decode(&user)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(user)
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var users []models.User
    collection := client.Database("crudexample").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        log.Fatal(err)
    }
    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        var user models.User
        cursor.Decode(&user)
        users = append(users, user)
    }
    json.NewEncoder(w).Encode(users)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var user models.User
    _ = json.NewDecoder(r.Body).Decode(&user)
    collection := client.Database("crudexample").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    filter := bson.M{"_id": id}
    update := bson.M{
        "$set": user,
    }
    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(user)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := r.URL.Query()
    id, _ := primitive.ObjectIDFromHex(params.Get("id"))
    collection := client.Database("crudexample").Collection("users")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    _, err := collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("User