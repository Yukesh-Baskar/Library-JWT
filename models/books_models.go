package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Book struct {
	ID       primitive.ObjectID `bson:"_id"`
	BookName string             `json:"book_name"`
	Author   string             `json:"author"`
	Category string             `json:"category"`
	Price    int                `json:"price"`
	Book_ID  string             `json:"book_id"`
}
