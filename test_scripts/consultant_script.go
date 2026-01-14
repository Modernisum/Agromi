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
	fmt.Println("🚀 Starting Consultant Module Tests...")
	time.Sleep(1 * time.Second)

	// 1. Register Consultant
	fmt.Println("\n[TEST 1] Registering New Consultant...")
	consultantID := apiPost("consultant/create", map[string]interface{}{
		"name":             "Dr. A. Kumar",
		"type":             "Doctor",
		"phone":            "9988776655",
		"address":          "New Delhi",
		"qualification":    []string{"MBBS", "PhD in Agriculture"},
		"experience":       12,
		"consultation_fee": 500,
		"timing":           "10:00 AM - 6:00 PM",
	})

	// 2. Admin Verifies Consultant
	if consultantID != "" {
		fmt.Println("\n[TEST 2] Admin Verifying Consultant...")
		apiPut("admin/finance/verify", map[string]interface{}{
			"id":          consultantID,
			"type":        "consultant",
			"is_verified": true,
		})
	}

	// 3. Admin Blocks Consultant
	if consultantID != "" {
		fmt.Println("\n[TEST 3] Admin Blocking Consultant...")
		// Use empty body for query param based endpoint if possible, but our PUT might need empty json
		url := fmt.Sprintf("%s/admin/consultant/manage/block/%s?action=block", baseURL, consultantID)
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte("{}")))
		client := &http.Client{}
		resp, _ := client.Do(req)
		printResp(resp)
	}

	// 4. List Consultants (Verified Only)
	fmt.Println("\n[TEST 4] Listing Verified Consultants...")
	apiGet("consultant/list?verified_only=true")

	// 5. Consultant Analytics
	fmt.Println("\n[TEST 5] Fetching Analytics...")
	apiGet("admin/consultant/analytics")

	fmt.Println("\n✅ Consultant Tests Completed.")
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
	return printResp(resp)
}

func apiPut(endpoint string, payload map[string]interface{}) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
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

func printResp(resp *http.Response) string {
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(body))

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		if id, ok := result["id"].(string); ok {
			return id
		}
	}
	return ""
}
