package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Consultant Types
const (
	TypeDoctor             = "Doctor"
	TypeAwardedFarmer      = "Awarded Farmer"
	TypeGovernmentAgent    = "Government Agent"
	TypeMarketVendor       = "Market Vendor"
	TypeTechnologyEngineer = "Technology Engineer"
	TypeOther              = "Other"
)

// Verification Status
const (
	StatusPending    = "Pending"
	StatusVerified   = "Verified"
	StatusUnverified = "Unverified"
)

// Consultant Struct
type Consultant struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AuthTokenNum string             `json:"auth_token_num" bson:"auth_token_num"` // For consistent auth handling if needed

	// Personal Info
	Name    string `json:"name" bson:"name"`
	Age     int    `json:"age" bson:"age"`
	Phone   string `json:"phone" bson:"phone"`
	Address string `json:"address" bson:"address"`

	// Professional Profile
	Type          string   `json:"type" bson:"type"`
	Qualification []string `json:"qualification" bson:"qualification"` // Degrees/Certs
	Experience    int      `json:"experience" bson:"experience"`       // Years
	Achievements  []string `json:"achievements" bson:"achievements"`
	Position      string   `json:"position" bson:"position"` // Current role

	// Verification
	VerificationStatus string `json:"verification_status" bson:"verification_status"`

	// Fees & Rates (Points)
	ConsultationFee float64 `json:"consultation_fee" bson:"consultation_fee"` // Base fee
	VoiceCallRate   float64 `json:"voice_call_rate" bson:"voice_call_rate"`   // Per minute
	VideoCallRate   float64 `json:"video_call_rate" bson:"video_call_rate"`   // Per minute
	ChatRate        float64 `json:"chat_rate" bson:"chat_rate"`               // Per session/msg

	// Availability
	Timing string `json:"timing" bson:"timing"` // e.g., "10:00 AM - 5:00 PM"

	// Media
	ProfilePhotoURL  string   `json:"profile_photo_url" bson:"profile_photo_url"`
	GalleryPhotoURLs []string `json:"gallery_photo_urls" bson:"gallery_photo_urls"`
	VideoURLs        []string `json:"video_urls" bson:"video_urls"`

	// Stats
	Rating      float64 `json:"rating" bson:"rating"`
	ReviewCount int     `json:"review_count" bson:"review_count"`

	// System
	IsBlocked           bool       `json:"is_blocked" bson:"is_blocked"`
	CreatedAt           time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" bson:"updated_at"`
	DeletionScheduledAt *time.Time `json:"deletion_scheduled_at,omitempty" bson:"deletion_scheduled_at,omitempty"`
}
