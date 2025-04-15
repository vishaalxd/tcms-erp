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

func CreateStoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var store models.Store
	_ = json.NewDecoder(r.Body).Decode(&store)
	collection := config.Client.Database("customer_vendor_api").Collection("stores")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, store)
	json.NewEncoder(w).Encode(result)
}

func GetStoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := r.URL.Query()
	id, _ := primitive.ObjectIDFromHex(params.Get("id"))
	var store models.Store
	collection := config.Client.Database("customer_vendor_api").Collection("stores")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, models.Store{ID: id}).Decode(&store)
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(store)
}

func GetStoresHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var stores []models.Store
	collection := config.Client.Database("customer_vendor_api").Collection("stores")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var store models.Store
		cursor.Decode(&store)
		stores = append(stores, store)
	}
	json.NewEncoder(w).Encode(stores)
}

func UpdateStoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var store models.Store
	_ = json.NewDecoder(r.Body).Decode(&store)
	collection := config.Client.Database("customer_vendor_api").Collection("stores")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	params := r.URL.Query()
	id, _ := primitive.ObjectIDFromHex(params.Get("id"))
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": store,
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(store)
}

func DeleteStoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := r.URL.Query()
	id, _ := primitive.ObjectIDFromHex(params.Get("id"))
	collection := config.Client.Database("customer_vendor_api").Collection("stores")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode("Store deleted")
}
