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
