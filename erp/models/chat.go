package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FromUserID primitive.ObjectID `bson:"from_user_id,omitempty" json:"from_user_id,omitempty"`
	ToAdmin    bool               `bson:"to_admin,omitempty" json:"to_admin,omitempty"`
	Content    string             `bson:"content" json:"content"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
}

type Chat struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID   primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	AdminID  primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	Messages []Message          `bson:"messages" json:"messages"`
}

type BroadcastMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AdminID   primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id,omitempty"`
	Content   string             `bson:"content" json:"content"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
