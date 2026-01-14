package community_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post Structure
type Post struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID     primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	SenderName   string             `bson:"sender_name" json:"sender_name"`
	SenderRating float64            `bson:"sender_rating" json:"sender_rating"` // Cached rating of sender

	Content  string   `bson:"content" json:"content"`
	MediaURL string   `bson:"media_url,omitempty" json:"media_url,omitempty"`
	Tags     []string `bson:"tags,omitempty" json:"tags,omitempty"`

	Location *GeoJSON `bson:"location,omitempty" json:"location,omitempty"`

	LikesCount int     `bson:"likes_count" json:"likes_count"`
	Score      float64 `bson:"score,omitempty" json:"score,omitempty"` // Computed score for feed

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// GeoJSON for 2dsphere index (Location)
type GeoJSON struct {
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [lon, lat]
}
