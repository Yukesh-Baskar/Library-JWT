package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/library-management-system/library/database"
	"github.com/library-management-system/library/helpers"
	"github.com/library-management-system/library/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection("users")
var validate = validator.New()

func SignUp(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusBadRequest}.HandleError())
		return
	}

	err := validate.Struct(user)
	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusBadRequest}.HandleError())
		return
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError}.HandleError())
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: "User with this email already exist!", Status: http.StatusBadRequest}.HandleError())
		return
	}

	count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError}.HandleError())
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: "User with this phone number already exist!", Status: http.StatusBadRequest}.HandleError())
		return
	}

	user.CreatedAt, _ = time.Parse(time.ANSIC, time.Now().Format(time.ANSIC))
	user.UpdatedAt, _ = time.Parse(time.ANSIC, time.Now().Format(time.ANSIC))
	user.Password, err = helpers.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: fmt.Sprintf("Error occured while hashing password: %s", err.Error()), Status: http.StatusInternalServerError}.HandleError())
		return
	}
	user.ID = primitive.NewObjectID()
	user.User_ID = user.ID.Hex()
	token, err := helpers.GenerateToken(user.Email, user.FirstName, user.LastName, user.User_ID, user.User_Type, 1)
	user.Refresh_Token = token[0]
	if err != nil {
		c.JSON(http.StatusInternalServerError, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError})
		return
	}
	res, err := userCollection.InsertOne(ctx, user)

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError}.HandleError())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"InsertedID": res.InsertedID,
	})
	return
}

func Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user *models.User
	var dbUser *models.User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusBadRequest}.HandleError())
		return
	}

	if user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: "Email or password is empty", Status: http.StatusBadRequest}.HandleError())
		return
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError}.HandleError())
		return
	}

	if count == 0 {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: "User with this email not exist!", Status: http.StatusNotFound}.HandleError())
		return
	}

	err = userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusInternalServerError}.HandleError())
		return
	}

	if !helpers.ValidatePassword(dbUser.Password, user.Password) {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: "Invalid password", Status: http.StatusBadRequest}.HandleError())
		return
	}

	token, err := helpers.GenerateToken(dbUser.Email, dbUser.FirstName, dbUser.LastName, dbUser.User_ID, dbUser.User_Type, 2)

	if err != nil {
		c.JSON(http.StatusBadGateway, helpers.Errors{Message: err.Error(), Status: http.StatusBadRequest}.HandleError())
		return
	}
	c.SetCookie("refreshToken", token[1], 86400*7, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{
		"accessToken": token[0],
	})
}

func AddBook(c *gin.Context) {
	token := c.GetString("token")
	fmt.Println(token)
	claims, err := helpers.ValidateToken(token, false)

	if err != nil {
		c.JSON(err.Status, err)
		return
	}
	fmt.Println("claims", claims)
}

func RefreshToken(c *gin.Context) {
	token := c.GetString("token")
	refreshToken := c.GetString("refreshToken")

	_, err := helpers.ValidateToken(refreshToken, false)

	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	claims, err := helpers.ValidateToken(token, true)

	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	tokens, tErr := helpers.GenerateToken(claims.Email, claims.FirstName, claims.LastName, claims.Uid, claims.User_Type, 1)

	if err != nil {
		c.JSON(err.Status, tErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": tokens[0],
	})

}
