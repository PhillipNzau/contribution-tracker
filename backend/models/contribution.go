package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contribution struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventID         primitive.ObjectID `bson:"event_id" json:"event_id"`
	ContributorName string             `bson:"contributor_name" json:"contributor_name"`
	ContributorContact string          `bson:"contributor_contact,omitempty" json:"contributor_contact,omitempty"`
	Amount          float64            `bson:"amount" json:"amount"`
	Currency        string             `bson:"currency" json:"currency"`
	Method          string             `bson:"method" json:"method"` // MPESA, STRIPE, CASH
	PaymentRef      string             `bson:"payment_reference,omitempty" json:"payment_reference,omitempty"`
	Status          string             `bson:"status" json:"status"` // PENDING, CONFIRMED, FAILED, RECONCILED
	ReceiptURL      string             `bson:"receipt_url,omitempty" json:"receipt_url,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
