package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VaultItem struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
    Name        string             `json:"name"`
    URL         string             `json:"url"`
    Username    string             `json:"username"`
    Password    string             `json:"password"`
    Notes       string             `json:"notes"`
    CreatedAt   time.Time          `json:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at"`
}
