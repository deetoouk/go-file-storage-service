package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// File struct is a description of the File data
type File struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	FileID      primitive.ObjectID `json:"file_id,omitempty" bson:"file_id,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
}
