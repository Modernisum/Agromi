package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "http://localhost:8080/api"

func main() {
	fmt.Println("🚀 Starting Social Module Tests...")
	time.Sleep(1 * time.Second)

	// 1. Follow User
	fmt.Println("\n[TEST 1] Following User...")
	// Assuming IDs from previous runs or created fresh.
	// For automation, we typically need real IDs.
	// Let's create two users first or assume they exist.
	// HACK: ID simulation for test script might fail if IDs invalid.
	// For robust test, we should Register -> Get ID.
	// Skipping registration for brevity, assuming standard flow works.
	// Will use placeholder valid MongoIDs for test structure if this was unit test,
	// but for integration, we'll try to use the ones from previous log if possible or create new.

	// Create User A (Follower)
	followerID := registerUser("FollowerUser", "9111111111")
	// Create User B (Target)
	targetID := registerUser("TargetUser", "9222222222")
	fmt.Printf("DEBUG: FollowerID=%s, TargetID=%s\n", followerID, targetID)

	if followerID != "" && targetID != "" {
		apiPost("social/follow", map[string]interface{}{
			"follower_id": followerID,
			"followee_id": targetID,
		})
	}

	// 2. Post Comment
	fmt.Println("\n[TEST 2] Posting Comment...")
	apiPost("social/comment/create", map[string]interface{}{
		"target_id":   targetID,
		"sender_id":   followerID,
		"sender_name": "FollowerUser",
		"text":        "Nice profile!",
		"owner_id":    targetID,
	})

	// 3. Like Post/User
	fmt.Println("\n[TEST 3] Liking User...")
	apiPost("social/reaction/like", map[string]interface{}{
		"target_id": targetID,
		"sender_id": followerID,
		"action":    "like",
		"owner_id":  targetID,
	})

	// 4. Check Notifications
	fmt.Println("\n[TEST 4] Checking Notifications...")
	apiGet(fmt.Sprintf("social/notification/list?user_id=%s", targetID))

	fmt.Println("\n✅ Social Tests Completed.")
}

func registerUser(name, phone string) string {
	url := fmt.Sprintf("%s/auth/register", baseURL)
	payload := map[string]interface{}{
		"name":              name,
		"phone":             phone,
		"user_type":         "farmer",
		"profile_photo_url": "http://example.com/photo.jpg",
		"regional_language": "English",
		"crops": []map[string]string{
			{"name": "Wheat", "area": "2 acres", "age": "3 months"},
		},
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("DEBUG REGISTER RESP: %s\n", string(body))
	if id, ok := result["user_id"].(string); ok {
		return id
	} // Verify key structure
	// Try login if register says exists
	if msg, ok := result["error"].(string); ok && msg == "User already exists" {
		return loginUser(phone)
	}
	// If other error or success structure different?
	// Check key "id" instead of "user_id" for register response?
	if id, ok := result["id"].(string); ok {
		return id
	} // Some APIs return "id"
	if id, ok := result["_id"].(string); ok {
		return id
	}

	fmt.Printf("Register Failed for %s: %s\n", phone, string(body))
	return loginUser(phone)
}

func loginUser(phone string) string {
	url := fmt.Sprintf("%s/auth/login", baseURL)
	payload := map[string]interface{}{"phone": phone, "firebase_token": "dummy_token"}
	jsonData, _ := json.Marshal(payload)
	resp, _ := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("DEBUG LOGIN RESP: %s\n", string(body))

	// Login returns "user_id"
	if id, ok := result["user_id"].(string); ok {
		return id
	}
	// Check for "id" just in case
	if id, ok := result["id"].(string); ok {
		return id
	}

	fmt.Printf("Login Failed for %s: %s\n", phone, string(body))
	return ""
}

func apiPost(endpoint string, payload map[string]interface{}) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	printResp(resp)
}

func apiGet(endpoint string) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	printResp(resp)
}

func printResp(resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(body))
}
