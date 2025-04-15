package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Feed struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt int64              `bson:"created_at" json:"created_at"`
}
