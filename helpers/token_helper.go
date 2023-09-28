package helpers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/library-management-system/library/database"
	"go.mongodb.org/mongo-driver/mongo"
)

type JWTSignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	User_Type string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection("users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateToken(email, firstName, lastName, uid, userType string, flag int) ([]string, error) {
	refreshClaims := &JWTSignedDetails{
		StandardClaims: jwt.StandardClaims{Issuer: "Library",
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix()},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return nil, fmt.Errorf("error occured while generating refresh token: %v \n", err)
	}
	if flag == 1 {
		return []string{refreshToken}, nil
	}
	claims := &JWTSignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Uid:       uid,
		User_Type: userType,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "Library",
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return nil, fmt.Errorf("error occured while generating token: %v \n", err)
	}
	return []string{token, refreshToken}, nil
}

func ValidateToken(signedToken string, isRefresh bool) (*JWTSignedDetails, *Errors) {
	token, err := jwt.ParseWithClaims(signedToken, &JWTSignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil && !isRefresh {
		if err == jwt.ErrSignatureInvalid {
			return nil, &Errors{
				Message: fmt.Sprintf("Invalid signature: %s", err.Error()),
				Status:  http.StatusUnauthorized,
			}
		}
		return nil, &Errors{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		}
	}

	claims, ok := token.Claims.(*JWTSignedDetails)

	if !ok {
		return nil, &Errors{
			Message: "Invalid token!",
			Status:  http.StatusBadRequest,
		}
	}

	if isRefresh {
		if claims.ExpiresAt > time.Now().Local().Unix() {
			return nil, &Errors{
				Message: "Token not expired yet!",
				Status:  http.StatusBadRequest,
			}
		}
		// Refresh can be only done within 2 minutes
		if claims.ExpiresAt < (time.Now().Local().Unix() - 60) {
			return nil, &Errors{
				Message: "Refresh exceeds 2 minutes!",
				Status:  http.StatusBadRequest,
			}
		}
	} else {
		if claims.ExpiresAt < time.Now().Local().Unix() {
			return nil, &Errors{
				Message: "Token expired!",
				Status:  http.StatusBadRequest,
			}
		}

	}
	return claims, nil
}
