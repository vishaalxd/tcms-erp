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
