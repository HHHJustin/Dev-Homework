package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("TodoListSecretKey")

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func ParseTokenAndGetUserID(tokenString string) (uint, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, errors.New("token expired")
		}
		return 0, errors.New("invalid token")
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["userid"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID in token")
	}

	return uint(userID), nil
}

// Function to create JWT tokens with claims
func CreateToken(username string, user_id int) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    username,                         // Subject (user identifier)
		"iss":    "todo-app",                       // Issuer
		"aud":    "user",                           // Audience (user role)
		"exp":    time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat":    time.Now().Unix(),                // Issued at
		"userid": user_id,
	})
	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	// Print information about the created token
	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	// Parse the token with the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	// Check for verification errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return token, nil
}
