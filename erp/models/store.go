package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Store struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name string             `bson:"name" json:"name"`
}
