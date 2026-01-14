package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductType constants
const (
	TypeBuy  = "buy" // Deprecated for new creations, but kept for legacy
	TypeRent = "rent"
	TypeSell = "sell"
)

// Scoring Weights
const (
	WeightRelevance = 0.4
	WeightDistance  = 0.3
	WeightRating    = 0.2
	WeightFreshness = 0.1
)

type GeoLocation struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [longitude, latitude]
}

type Specification struct {
	Type  string `json:"type" bson:"type"`   // e.g., "Engine", "Dimension"
	Name  string `json:"name" bson:"name"`   // e.g., "V8", "10x10"
	Value string `json:"value" bson:"value"` // Optional extra value
}

type Tag struct {
	Type string `json:"type" bson:"type"` // e.g., "Company", "Model", "Category"
	Name string `json:"name" bson:"name"` // e.g., "Mahindra", "2018", "Tractor"
}

type Product struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type        string             `json:"type" bson:"type"` // rent, sell
	Category    string             `json:"category" bson:"category"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Price       float64            `json:"price" bson:"price"`
	ImageURL    string             `json:"image_url" bson:"image_url"`

	// New Fields
	Quantity float64 `json:"quantity" bson:"quantity"`
	Unit     string  `json:"unit" bson:"unit"` // e.g., "kg", "liters", "units"

	Specifications []Specification `json:"specifications" bson:"specifications"`
	Tags           []Tag           `json:"tags" bson:"tags"`

	Location GeoLocation `json:"location" bson:"location"`
	Address  string      `json:"address" bson:"address"`

	OwnerID     primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	Rating      float64            `json:"rating" bson:"rating"`
	ReviewCount int                `json:"review_count" bson:"review_count"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`

	Priority    int  `json:"priority" bson:"priority"` // Used as Score
	IsSponsored bool `json:"is_sponsored" bson:"is_sponsored"`
	IsBlocked   bool `json:"is_blocked" bson:"is_blocked"`

	Comments []Comment `json:"comments" bson:"comments"`
}

type Comment struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName  string             `json:"user_name" bson:"user_name"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
