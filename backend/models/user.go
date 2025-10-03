package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	Role      	 string             `bson:"role" json:"role"`           // e.g., host, manager, cleaner
	Phone     	 string             `bson:"phone,omitempty" json:"phone,omitempty"`
	RefreshToken string             `bson:"refresh_token,omitempty" json:"-"`
	OTP          string             `bson:"otp,omitempty" json:"-"`
	OTPExpiry    time.Time          `bson:"otp_expiry,omitempty" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}