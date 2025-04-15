package handlers

import (
	"adonai-api/config"
	"adonai-api/models"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OTPRequest struct {
	PhoneNumber string `json:"phone_number"`
	OTP         string `json:"otp"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"` // Add Role here
	jwt.StandardClaims
}

func generateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func sendSMS(to string, body string) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetBody(body)

	_, err := client.Api.CreateMessage(params)
	return err
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	collection := config.Client.Database("customer_vendor_api").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func RequestOTPHandler(w http.ResponseWriter, r *http.Request) {
	var phoneRequest OTPRequest
	err := json.NewDecoder(r.Body).Decode(&phoneRequest)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	collection := config.Client.Database("customer_vendor_api").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	var user models.User
	err = collection.FindOne(ctx, bson.M{"phone_number": phoneRequest.PhoneNumber}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	otp := generateOTP()
	expirationTime := time.Now().Add(5 * time.Minute).Unix() // OTP expires in 5 minutes
	user.OTP = otp
	user.OTPExpiresAt = expirationTime

	_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"otp":            user.OTP,
			"otp_expires_at": user.OTPExpiresAt,
		},
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = sendSMS(phoneRequest.PhoneNumber, "Your OTP is: "+otp)
	if err != nil {
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OTP sent successfully")
}

func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var otpRequest OTPRequest
	err := json.NewDecoder(r.Body).Decode(&otpRequest)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	collection := config.Client.Database("customer_vendor_api").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	var user models.User
	err = collection.FindOne(ctx, bson.M{"phone_number": otpRequest.PhoneNumber}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.OTP != otpRequest.OTP || time.Now().Unix() > user.OTPExpiresAt {
		http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	// Generate JWT Token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	fmt.Fprintf(w, "Login successful")
}
