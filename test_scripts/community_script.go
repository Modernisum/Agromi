package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "http://localhost:8080/api/community"

func main() {
	fmt.Println("🚀 Starting Community Module Tests...")
	time.Sleep(1 * time.Second)

	// 1. Create Post
	fmt.Println("\n[TEST 1] Creating Community Post...")
	postID := createPost()
	if postID == "" {
		fmt.Println("❌ Failed to create post (Mock ID used)")
		return
	}
	fmt.Printf("✅ Post Created. ID: %s\n", postID)

	// 2. Get Feed (Scored)
	fmt.Println("\n[TEST 2] Fetching Feed...")
	// Providing Lat/Lon to test distance scoring
	url := fmt.Sprintf("%s/feed?lat=12.9716&lon=77.5946&query=", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Feed Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(body))

	fmt.Println("\n✅ Community Tests Completed.")
}

func createPost() string {
	payload := map[string]interface{}{
		"sender_id":     "507f1f77bcf86cd799439011", // Mock ObjectID
		"sender_name":   "Community Farmer",
		"content":       "Hello Community! This is a relevant message.",
		"lat":           12.9716, // Bangalore
		"lon":           77.5946,
		"sender_rating": 4.5,
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}
