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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection = database.OpenCollection("users")
var bookCollection *mongo.Collection = database.OpenCollection("books")
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
	fmt.Println("wesgews")
	token := c.GetString("token")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	claims, err := helpers.ValidateToken(token, false)

	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	if claims.User_Type != "ADMIN" {
		c.JSON(http.StatusInternalServerError, helpers.Errors{Message: "User cannot access this resource!", Status: http.StatusInternalServerError}.HandleError())
		return
	}

	var book *models.Book

	if err := c.BindJSON(&book); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	filter := bson.D{{Key: "category", Value: book.Category}}
	count, cerr := bookCollection.CountDocuments(ctx, filter)

	if cerr != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Message,
		})
		return
	}
	var id any
	if count > 0 {
		update := bson.M{"$inc": bson.M{"total_count": 1}}
		options := options.Update().SetUpsert(true)
		ressss, cerr := bookCollection.UpdateOne(ctx, filter, update, options)
		if cerr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Message,
			})
			return
		}
		fmt.Println("r", ressss)
		c.JSON(http.StatusAccepted, ressss)
	} else {
		book.ID = primitive.NewObjectID()
		book.Book_ID = book.ID.Hex()
		id, cerr = bookCollection.InsertOne(ctx, book)
		if cerr != nil {
			c.JSON(http.StatusInternalServerError, helpers.Errors{Message: cerr.Error(), Status: http.StatusInternalServerError})
			return
		}
	}
	c.JSON(http.StatusOK, id)
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

func BuyBook(c *gin.Context) {
	token := c.GetString("token")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	claims, err := helpers.ValidateToken(token, false)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	var book models.Book
	if bErr := c.Bind(&book); bErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, bErr)
		return
	}

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "category", Value: book.Category},
			{Key: "total_count", Value: bson.D{{Key: "$gt", Value: 0}}},
		}},
	}

	cur, cerr := bookCollection.Aggregate(ctx, mongo.Pipeline{matchStage})
	if cerr != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var res []bson.M
	if err := cur.All(ctx, &res); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	if len(res) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No books available in this category",
		})
		return
	}

	// Define the filter for the book update
	bookFilter := bson.D{
		{Key: "category", Value: book.Category},
		{Key: "total_count", Value: bson.D{{Key: "$gt", Value: 0}}},
	}

	// Define the update for the book
	bookUpdate := bson.M{"$inc": bson.M{"total_count": -1}}

	updateRes, cerr := bookCollection.UpdateOne(ctx, bookFilter, bookUpdate)

	if cerr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": cerr.Error(),
		})
		return
	}

	if updateRes.ModifiedCount == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "No matching document found or document not updated.",
		})
		return
	}

	userFilter := bson.D{{Key: "user_id", Value: claims.Uid}}
	userUpdate := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "books", Value: bson.D{{Key: "$each", Value: res}}},
		}},
	}

	updateRes, cerr = userCollection.UpdateOne(ctx, userFilter, userUpdate)

	if cerr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": cerr.Error(),
		})
		return
	}

	if updateRes.ModifiedCount == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "No matching document found or document not updated.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Book bought successfully.",
	})
}
