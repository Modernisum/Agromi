package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RunDebugDB() {
	uri := "mongodb+srv://modernisum_db_user:Sg2FVq9Jwu5B7Eku@cluster0.pc5owye.mongodb.net/?appName=Cluster0"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("modernisum_db").Collection("users")

	// Print all users (not just farmers to be safe)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total users: %d\n", len(results))
	for i, doc := range results {
		fmt.Printf("[%d] Name: '%v' | Phone: '%v' | Type: '%v'\n", i, doc["name"], doc["phone"], doc["user_type"])
	}
}
