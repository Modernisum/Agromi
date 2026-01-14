package social_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment Structure
type Comment struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetID   primitive.ObjectID `bson:"target_id" json:"target_id"` // Product or User ID
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	SenderName string             `bson:"sender_name" json:"sender_name"`
	Text       string             `bson:"text" json:"text"`
	MediaURL   string             `bson:"media_url,omitempty" json:"media_url,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// Like/Dislike Structure
type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetID  primitive.ObjectID `bson:"target_id" json:"target_id"`
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Action    string             `bson:"action" json:"action"` // "like", "dislike"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Review Structure
type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetID  primitive.ObjectID `bson:"target_id" json:"target_id"` // Consultant or Product ID
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Rating    float64            `bson:"rating" json:"rating"` // 1-5
	Text      string             `bson:"text" json:"text"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Follow Structure
type Follow struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FollowerID primitive.ObjectID `bson:"follower_id" json:"follower_id"`
	FolloweeID primitive.ObjectID `bson:"followee_id" json:"followee_id"` // User being followed
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

// Notification Structure
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RecipientID primitive.ObjectID `bson:"recipient_id" json:"recipient_id"`
	Type        string             `bson:"type" json:"type"` // "comment", "like", "follow", "new_post"
	Message     string             `bson:"message" json:"message"`
	RelatedID   primitive.ObjectID `bson:"related_id,omitempty" json:"related_id,omitempty"` // ID of comment/post/user
	IsRead      bool               `bson:"is_read" json:"is_read"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
