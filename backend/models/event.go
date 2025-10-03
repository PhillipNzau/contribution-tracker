package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"` // Organizer
	Title        string             `bson:"title" json:"title"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	Location     string             `bson:"location,omitempty" json:"location,omitempty"`
	TargetAmount float64            `bson:"target_amount,omitempty" json:"target_amount,omitempty"`
	Deadline     *time.Time         `bson:"deadline,omitempty" json:"deadline,omitempty"`
	Status       string             `bson:"status" json:"status"` // ACTIVE, CLOSED, ARCHIVED
	Images       []string            `bson:"images" json:"images"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
