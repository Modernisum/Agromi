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
	fmt.Println("🚀 Starting Automated Marketplace Tests...")
	time.Sleep(1 * time.Second)

	// 1. Farmer Creates Sell Listing (Buffalo)
	fmt.Println("\n[TEST 1] Creating Farmer Sell Listing (Buffalo)...")
	_ = createListing("market/sell/create", map[string]interface{}{
		"type":        "sell",
		"category":    "Animals",
		"name":        "Murrah Buffalo",
		"description": "Healthy Murrah buffalo, 2nd lactation",
		"price":       85000,
		"quantity":    1,
		"unit":        "cattle",
		"location": map[string]interface{}{
			"type":        "Point",
			"coordinates": []float64{77.2090, 28.6139},
		},
		"specifications": []map[string]interface{}{
			{"type": "Physical", "name": "Lactation", "value": "2nd"},
		},
		"tags": []map[string]interface{}{
			{"type": "Animal", "name": "Buffalo"},
		},
	})

	// 2. Admin Adds Buy Item (Gir Cow)
	fmt.Println("\n[TEST 2] Admin Adding Buy Item (Gir Cow)...")
	cowID := createListing("admin/market/buy/add", map[string]interface{}{
		"category":    "Animals",
		"name":        "Gir Cow (Certified)",
		"description": "Purebred Gir cow",
		"price":       65000,
		"quantity":    5,
		"unit":        "cattle",
		"location": map[string]interface{}{
			"type":        "Point",
			"coordinates": []float64{72.8777, 19.0760},
		},
		"tags": []map[string]interface{}{
			{"type": "Animal", "name": "Cow"},
		},
	})

	// 3. Admin Updates Score (Finance)
	if cowID != "" {
		fmt.Println("\n[TEST 3] Admin Updating Product Score...")
		updateScore(cowID, 150)
	}

	// 4. List All Sell Items
	fmt.Println("\n[TEST 4] Listing All Sell Items...")
	listItems("market/sell/list")

	// 5. List All Buy Items
	fmt.Println("\n[TEST 5] Listing All Buy Items...")
	listItems("market/buy/list")

	fmt.Println("\n✅ All Tests Completed.")
}

func createListing(endpoint string, payload map[string]interface{}) string {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

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

func updateScore(productID string, score int) {
	url := fmt.Sprintf("%s/admin/finance/sponsor/score", baseURL)
	payload := map[string]interface{}{
		"product_id": productID,
		"score":      score,
	}
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

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(body))
}

func listItems(endpoint string) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// Truncate long output for display
	output := string(body)
	if len(output) > 500 {
		output = output[:500] + "... (truncated)"
	}
	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, output)
}
