package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"my_store_app/database"
	"my_store_app/models"
	request "my_store_app/requests"
	"my_store_app/response"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Generating our access tokens
// Secret key used to sign tokens (in production, keep it secret and secure)
var jwtKey = []byte("your_secret_key")

// UserClaims struct to hold custom claims for the token
type UserClaims struct {
    Email string `json:"email"`
    jwt.RegisteredClaims
}

// Function to generate JWT token
func generateJWT(user models.User) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

    claims := &UserClaims{
        Email: user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

// Login function to handle the request
func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }
    defer r.Body.Close()

    var signInReq request.SignInRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&signInReq); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Initialize the validator and validate the request data
    validate := validator.New()
    err := validate.Struct(signInReq)
    if err != nil {
        if errs, ok := err.(validator.ValidationErrors); ok {
            var errMessages []string
            for _, e := range errs {
                errMessages = append(errMessages, e.Error())
            }
            http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
            return
        }
        http.Error(w, "Validation failed...", http.StatusInternalServerError)
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Assuming `database` and `Client` are predefined in your package
    collection := database.OpenCollection(database.Client, "users")
       // Find user by email
       var user models.User
       err = collection.FindOne(ctx, bson.M{"username": signInReq.Username}).Decode(&user)
       if err != nil {
           if err == mongo.ErrNoDocuments {
            err_resp:=response.ErrorResponse{
                OK:false,
                Status:"failed",
                Message: "Invalid username or password" ,
            }
            errJSON, _ := json.Marshal(err_resp)
               http.Error(w, string(errJSON), http.StatusUnauthorized)
               return
           }
           http.Error(w, "Failed to find user", http.StatusInternalServerError)
           return
       }
   
       // Compare the password with the hashed password in the database
       err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(signInReq.Password))
       if err != nil {
        err_resp:=response.ErrorResponse{
            OK:false,
            Status:"failed",
            Message: "Invalid username or password" ,
        }
        errJSON, _ := json.Marshal(err_resp)
           http.Error(w, string(errJSON), http.StatusUnauthorized)
           return
       }
   
       // If you want to generate a token (e.g., JWT), do it here and include it in the response
        token, err := generateJWT(user)
        if err != nil {
            http.Error(w, "Failed to generate token", http.StatusInternalServerError)
            return
        }
   
       // Return success response (including token if generated)
       w.Header().Set("Content-Type", "application/json")
       json.NewEncoder(w).Encode(map[string]interface{}{
           "ok":     true,
           "status": "success",
           "message": "User signed in successfully",
           "users":user,
            "token":  token, // Include the token if generated
       })
   
}

// SignUp function to handle the request
func SignUp(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }
    defer r.Body.Close()

    var signUpReq request.SignUpRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&signUpReq); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Initialize the validator and validate the request data
    validate := validator.New()
    err := validate.Struct(signUpReq)
    if err != nil {
        if errs, ok := err.(validator.ValidationErrors); ok {
            var errMessages []string
            for _, e := range errs {
                errMessages = append(errMessages, e.Error())
            }
            http.Error(w, strings.Join(errMessages, ", "), http.StatusBadRequest)
            return
        }
        http.Error(w, "Validation failed", http.StatusInternalServerError)
        return
    }

    // Assuming newUser model matches the fields from signUpReq
    newUser := models.User{
        Username: signUpReq.Username,
        Name: signUpReq.Name,
        Email: signUpReq.Email,
        Password: signUpReq.Password, // will be hashed below
        ID: uuid.NewString(), // Generate unique ID
        CreatedAt: time.Now(), // Set created_at and updated_at
        UpdatedAt: time.Now(),
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    newUser.Password = string(hashedPassword)
    //let have the logic to create the new users
    //lets first open the collection we need to insert the user in
    collection:= database.OpenCollection(database.Client,"users")
    ctx,cancel :=context.WithTimeout(context.Background(),10*time.Second)
    defer cancel()
    //let's now insert the user
    _,err=collection.InsertOne(ctx,newUser)
    if err != nil {
        err_response := response.ErrorResponse{
            OK: false,
            Message: err.Error(),
            Status: "failed",
        }
        json.NewEncoder(w).Encode(err_response)
    }


    // Properly respond with the created user, omitting sensitive data like password
    responseUser := newUser
    responseUser.Password = "" // Clear password before sending response
    w.Header().Set("Content-Type", "application/json")
    response:=response.Response{
        OK: true,
        Status: "success",
        Message: fmt.Sprintf("%s added successfully", newUser.Name),
        User: responseUser,
    }
    json.NewEncoder(w).Encode(response)

}

// GetUser function to handle the request
func GetUser(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Assuming `database` and `Client` are predefined in your package
    collection := database.OpenCollection(database.Client, "users")

   // Get id from the query string
   queryValues := r.URL.Query()
   userID := queryValues.Get("id")

    // Create a filter to find the user by ID
    filter := bson.M{"id": userID}

    // Finding the user document with the given ID
    var user bson.M
    err := collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "No user found with given ID", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    delete(user, "password") // Remove the password field
   var res_user [ ]bson.M 
    res_user=append(res_user,user)

    resResult := map[string]interface{}{
        "ok":     true,
        "status": "success",
        "user":   res_user,
    }

    // Set response header
    w.Header().Set("Content-Type", "application/json")

    // Encoding the result to JSON and sending the response
    if err := json.NewEncoder(w).Encode(resResult); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Println("Endpoint Hit: GetUser")
}

// GetUsers function to handle the request
func GetUsers(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel() // This ensures that context cancellation is always called

    // Assuming `database` and `Client` are predefined in your package
    collection := database.OpenCollection(database.Client, "users")

    // Finding all documents in the collection
    cur, err := collection.Find(ctx, bson.D{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(ctx)

    // Slice to hold all decoded documents
    var users []bson.M

    // Decoding documents into users slice
    for cur.Next(ctx) {
        var user bson.M
        if err := cur.Decode(&user); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        delete(user, "password")
        users = append(users, user)
    }
     res_result  := map[string]interface{}{
        "ok": true,
        "status": "success",
        "users": users,
    }

    if err := cur.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Set response header
    w.Header().Set("Content-Type", "application/json")

    // Encoding all users to JSON and sending the response
    if err := json.NewEncoder(w).Encode(res_result); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Println("Endpoint Hit: GetUsers")
}
