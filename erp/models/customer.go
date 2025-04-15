package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Customer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	FirstName string             `bson:"first_name" json:"first_name"`
	LastName  string             `bson:"last_name" json:"last_name"`
	StoreID   primitive.ObjectID `bson:"store_id,omitempty" json:"store_id,omitempty"`
}
