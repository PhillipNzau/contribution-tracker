package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title     string             `bson:"title" json:"title"`
	Message   string             `bson:"message" json:"message"`
	Read      bool               `bson:"read" json:"read"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
