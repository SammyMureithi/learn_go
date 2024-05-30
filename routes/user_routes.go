package routes

import (
	"my_store_app/controllers"
	"my_store_app/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

// UserRoutes function to initialize user routes
func UserRoutes(router *mux.Router) {
      // Apply middleware to the GET routes
      router.Handle("/users", middleware.JWTMiddleware(http.HandlerFunc(controllers.GetUsers))).Methods("GET")
      router.Handle("/user", middleware.JWTMiddleware(http.HandlerFunc(controllers.GetUser))).Methods("GET")
      
    router.HandleFunc("/users/signup", controllers.SignUp).Methods("POST")
    router.HandleFunc("/users/login", controllers.Login).Methods("POST")
}
