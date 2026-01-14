package chat_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message Structure
type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	ReceiverID primitive.ObjectID `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"` // For 1-on-1
	GroupID    primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`       // For Group Chat

	Content  string `bson:"content" json:"content"`
	MediaURL string `bson:"media_url,omitempty" json:"media_url,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// ChatGroup Structure
type ChatGroup struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string               `bson:"name" json:"name"`
	AdminID   primitive.ObjectID   `bson:"admin_id" json:"admin_id"`
	MemberIDs []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
}

// Conversation Config (Admin controlled)
const MaxMessagesPerChat = 500
