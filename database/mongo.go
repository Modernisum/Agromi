package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client       *mongo.Client
	clientOnce   sync.Once
	DatabaseName = "modernisum_db"
	UpsertOpt    = options.Update().SetUpsert(true)
)

// Connect initializes the MongoDB connection (Singleton)
func Connect() {
	clientOnce.Do(func() {
		// Use the provided connection string
		// Note: Ideally this should come from environment variables for security,
		// but using direct string as requested for this architecture.
		uri := "mongodb+srv://modernisum_db_user:Sg2FVq9Jwu5B7Eku@cluster0.pc5owye.mongodb.net/?appName=Cluster0"

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatal("Failed to create MongoDB client:", err)
		}

		// Ping the database to verify connection
		err = Client.Ping(ctx, nil)
		if err != nil {
			log.Fatal("Failed to ping MongoDB:", err)
		}

		fmt.Println("✅ Connected to MongoDB Atlas successfully!")
	})
}

// GetCollection is a helper to get a collection from the default database
func GetCollection(collectionName string) *mongo.Collection {
	if Client == nil {
		log.Fatal("Database client is not initialized. Call Connect() first.")
	}
	return Client.Database(DatabaseName).Collection(collectionName)
}
