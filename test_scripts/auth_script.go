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
	fmt.Println("🚀 Starting Unified Auth Tests...")
	time.Sleep(1 * time.Second)

	// 1. Register a Farmer (if not exists)
	fmt.Println("\n[TEST 1] Registering Farmer...")
	farmerPhone := "9000011111"
	apiPost("auth/register", map[string]interface{}{
		"name":      "Test Farmer",
		"phone":     farmerPhone,
		"user_type": "farmer",
	})

	// 2. Register a Consultant
	fmt.Println("\n[TEST 2] Registering Consultant...")
	consultantPhone := "9000022222"
	apiPost("consultant/create", map[string]interface{}{
		"name":  "Test Consultant",
		"phone": consultantPhone,
		"type":  "Doctor",
	})

	// 3. Login as Farmer
	fmt.Println("\n[TEST 3] Unified Login (Farmer)...")
	loginAndCheckType(farmerPhone, "farmer")

	// 4. Login as Consultant
	fmt.Println("\n[TEST 4] Unified Login (Consultant)...")
	loginAndCheckType(consultantPhone, "consultant")

	fmt.Println("\n✅ Unified Auth Tests Completed.")
}

func loginAndCheckType(phone, expectedType string) {
	url := fmt.Sprintf("%s/auth/login", baseURL)
	payload := map[string]interface{}{
		"phone":          phone,
		"firebase_token": "dummy_token",
	}
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Login Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		userType := result["user_type"].(string)
		if userType == expectedType {
			fmt.Printf("✅ Success: Logged in as %s (Type: %s)\n", phone, userType)
		} else {
			fmt.Printf("❌ Failed: Expected %s, Got %s\n", expectedType, userType)
		}
	} else {
		fmt.Printf("❌ Login Failed Status %s: %s\n", resp.Status, string(body))
	}
}

func apiPost(endpoint string, payload map[string]interface{}) string {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return ""
	}
	defer resp.Body.Close()
	return ""
}
