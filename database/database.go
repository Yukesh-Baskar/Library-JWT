package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ClientConnection = connectDB()

func connectDB() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error occured while loading env: %s \n", err.Error())
	}

	uri := os.Getenv("MONGO_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatalf("error occured while connecting new client: %s \n", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Microsecond)

	defer cancel()

	if err = client.Connect(ctx); err != nil {
		log.Fatalf("error occured while connecting DB: %s \n", err.Error())
	}
	fmt.Println("DB connected successfully")
	return client
}

func OpenCollection(collectionName string) *mongo.Collection {
	return ClientConnection.Database("library_management_system").Collection(collectionName)
}
