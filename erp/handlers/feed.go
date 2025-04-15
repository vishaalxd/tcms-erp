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
