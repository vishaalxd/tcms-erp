package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"password"`
	Role         string             `bson:"role" json:"role"` // "customer" or "vendor"
	PhoneNumber  string             `bson:"phone_number" json:"phone_number"`
	OTP          string             `bson:"otp" json:"otp"`
	OTPExpiresAt int64              `bson:"otp_expires_at" json:"otp_expires_at"`
}
