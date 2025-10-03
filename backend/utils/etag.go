package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateETag - make an ETag from object ID + updatedAt
func GenerateETag(id primitive.ObjectID, updatedAt time.Time) string {
	data := fmt.Sprintf("%s-%d", id.Hex(), updatedAt.UnixNano())
	hash := md5.Sum([]byte(data))
	return `"` + hex.EncodeToString(hash[:]) + `"`
}
