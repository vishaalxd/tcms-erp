package handlers

import (
	"adonai-api/config"
	"adonai-api/models"
	"context"
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

	collection := config.Client.Database("customer_vendor_api").Collection("orders")
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
