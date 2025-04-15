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

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	var msg models.Message
	_ = json.NewDecoder(r.Body).Decode(&msg)
	msg.Timestamp = time.Now()

	// collection := config.Client.Database("adonai-api").Collection("chats")
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
