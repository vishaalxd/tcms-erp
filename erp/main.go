package main

import (
	"adonai-api/config"
	"adonai-api/handlers"
	"adonai-api/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	config.InitEnv()
	config.ConnectDB()

	r := mux.NewRouter()

	// Authentication routes
	r.HandleFunc("/signup", handlers.SignUpHandler).Methods("POST")
	r.HandleFunc("/request-otp", handlers.RequestOTPHandler).Methods("POST")
	r.HandleFunc("/verify-otp", handlers.VerifyOTPHandler).Methods("POST")

	// Customer routes
	r.Handle("/customers", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetCustomersHandler))).Methods("GET")
	r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateCustomerHandler))).Methods("POST")
	r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetCustomerHandler))).Methods("GET")
	r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.UpdateCustomerHandler))).Methods("PUT")
	r.Handle("/customer", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.DeleteCustomerHandler))).Methods("DELETE")

	// Store routes
	r.Handle("/stores", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetStoresHandler))).Methods("GET")
	r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateStoreHandler))).Methods("POST")
	r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetStoreHandler))).Methods("GET")
	r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.UpdateStoreHandler))).Methods("PUT")
	r.Handle("/store", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.DeleteStoreHandler))).Methods("DELETE")

	// Order routes
	r.Handle("/orders", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetUserOrdersHandler))).Methods("GET")
	r.Handle("/order", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateOrderHandler))).Methods("POST")
	r.Handle("/cancel-order", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CancelOrderHandler))).Methods("PUT")
	r.Handle("/all-orders", middleware.JwtAuthMiddleware(middleware.RoleMiddleware("vendor")(http.HandlerFunc(handlers.GetAllOrdersHandler)))).Methods("GET")

	// Chat routes
	r.Handle("/send-message", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.SendMessageHandler))).Methods("POST")
	r.Handle("/chat-history", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetChatHistoryHandler))).Methods("GET")
	r.Handle("/broadcast", middleware.JwtAuthMiddleware(middleware.RoleMiddleware("vendor")(http.HandlerFunc(handlers.BroadcastMessageHandler)))).Methods("POST")

	// Feed routes
	r.Handle("/feeds", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.GetFeedsHandler))).Methods("GET")
	r.Handle("/feed", middleware.JwtAuthMiddleware(http.HandlerFunc(handlers.CreateFeedHandler))).Methods("POST")

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
