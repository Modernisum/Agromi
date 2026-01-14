package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var baseURL = "http://localhost:8080/api/chat"

func main() {
	fmt.Println("🚀 Starting Chat Module Tests...")
	time.Sleep(1 * time.Second)

	// Mock IDs
	userA := "507f1f77bcf86cd799439011"
	userB := "507f1f77bcf86cd799439012"
	var groupID string

	// 1. Send 1-on-1 Message
	fmt.Println("\n[TEST 1] Sending 1-on-1 Message...")
	msgID := sendMessage(userA, userB, "", "Hello Friend! Here is a photo", "http://photo.com/1.jpg")
	if msgID != "" {
		fmt.Printf("✅ Message Sent. ID: %s\n", msgID)
	} else {
		fmt.Println("❌ Failed to send message")
	}

	// 2. Create Group
	fmt.Println("\n[TEST 2] Creating Group...")
	groupID = createGroup("Farmers United", userA)
	if groupID != "" {
		fmt.Printf("✅ Group Created. ID: %s\n", groupID)
	}

	// 3. Join Group
	fmt.Println("\n[TEST 3] User B Joining Group...")
	joinGroup(groupID, userB)

	// 4. Send Group Message
	fmt.Println("\n[TEST 4] Sending Group Message...")
	grpMsgID := sendMessage(userB, "", groupID, "Hello Group!", "")
	if grpMsgID != "" {
		fmt.Printf("✅ Group Message Sent. ID: %s\n", grpMsgID)
	}

	// 5. Get History
	fmt.Println("\n[TEST 5] Fetching 1-on-1 History...")
	fetchHistory(userA, userB, false)

	fmt.Println("\n✅ Chat Tests Completed.")
}

func sendMessage(sender, receiver, group, content, media string) string {
	payload := map[string]interface{}{
		"sender_id": sender,
		"content":   content,
		"media_url": media,
	}
	if group != "" {
		payload["group_id"] = group
	} else {
		payload["receiver_id"] = receiver
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}

func createGroup(name, admin string) string {
	payload := map[string]interface{}{"name": name, "admin_id": admin}
	jsonData, _ := json.Marshal(payload)
	resp, _ := http.Post(baseURL+"/group/create", "application/json", bytes.NewBuffer(jsonData))
	defer resp.Body.Close()
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}

func joinGroup(group, user string) {
	payload := map[string]interface{}{"group_id": group, "user_id": user}
	jsonData, _ := json.Marshal(payload)
	http.Post(baseURL+"/group/join", "application/json", bytes.NewBuffer(jsonData))
}

func fetchHistory(user, other string, isGroup bool) {
	url := fmt.Sprintf("%s/history?user_id=%s&other_id=%s&is_group=%v", baseURL, user, other, isGroup)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
}
