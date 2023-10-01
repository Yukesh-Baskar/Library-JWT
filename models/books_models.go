package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Book struct {
	ID         primitive.ObjectID `bson:"_id"`
	BookName   string             `json:"book_name,omitempty"`
	Author     string             `json:"author,omitempty"`
	Category   string             `json:"category,omitempty"`
	TotalCount int                `json:"total_cont,omitempty"`
	Price      int                `json:"price,omitempty"`
	Book_ID    string             `json:"book_id,omitempty"`
}

type Books struct {
	Books []Book `bson:"total_books,omitempty" json:"total_books,omitempty"`
}
