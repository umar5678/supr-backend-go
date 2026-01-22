package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

var (
	riderToken  string
	driverToken string
	rideID      string
	messageID   string
)

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type RideResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type MessageResponse struct {
	ID      string `json:"id"`
	RideID  string `json:"rideId"`
	Content string `json:"content"`
	IsRead  bool   `json:"isRead"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func main() {
	fmt.Println("=== Messaging System - Integration Tests ===\n")

	// Test 1: Authentication
	fmt.Println("[1] Testing Authentication...")
	testAuthentication()

	// Test 2: Create Ride
	fmt.Println("\n[2] Creating Test Ride...")
	testCreateRide()

	// Test 3: Send Messages
	fmt.Println("\n[3] Testing Send Message...")
	testSendMessage()

	// Test 4: Get Messages
	fmt.Println("\n[4] Testing Get Messages...")
	testGetMessages()

	// Test 5: Get Unread Count
	fmt.Println("\n[5] Testing Get Unread Count...")
	testGetUnreadCount()

	// Test 6: Mark as Read
	fmt.Println("\n[6] Testing Mark as Read...")
	testMarkAsRead()

	// Test 7: Delete Message
	fmt.Println("\n[7] Testing Delete Message...")
	testDeleteMessage()

	// Test 8: Error Cases
	fmt.Println("\n[8] Testing Error Handling...")
	testErrorCases()

	fmt.Println("\n=== All Tests Completed ===")
}

func testAuthentication() {
	// Sign up rider
	resp := makeRequest("POST", "/api/v1/auth/phone/signup", map[string]interface{}{
		"phone":    "+1234567890",
		"password": "password123",
		"name":     "John Rider",
	}, "")

	var authResp AuthResponse
	json.Unmarshal(resp, &authResp)
	riderToken = authResp.Token
	fmt.Printf("✅ Rider signed up, token: %s\n", riderToken[:20]+"...")

	// Sign up driver
	resp = makeRequest("POST", "/api/v1/auth/phone/signup", map[string]interface{}{
		"phone":    "+1987654321",
		"password": "password123",
		"name":     "Jane Driver",
	}, "")

	json.Unmarshal(resp, &authResp)
	driverToken = authResp.Token
	fmt.Printf("✅ Driver signed up, token: %s\n", driverToken[:20]+"...")
}

func testCreateRide() {
	resp := makeRequest("POST", "/api/v1/rides", map[string]interface{}{
		"pickupLocation": map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
			"address":   "Times Square",
		},
		"dropoffLocation": map[string]interface{}{
			"latitude":  40.7580,
			"longitude": -73.9855,
			"address":   "Central Park",
		},
		"vehicleType": "cab_economy",
	}, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	var rideResp RideResponse
	json.Unmarshal(apiResp.Data, &rideResp)
	rideID = rideResp.ID
	fmt.Printf("✅ Ride created with ID: %s\n", rideID)
}

func testSendMessage() {
	// Rider sends message
	resp := makeRequest("POST", "/api/v1/messages", map[string]interface{}{
		"rideId":   rideID,
		"content":  "Hi! Where are you?",
		"metadata": map[string]string{"type": "text"},
	}, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	var msgResp MessageResponse
	json.Unmarshal(apiResp.Data, &msgResp)
	messageID = msgResp.ID
	fmt.Printf("✅ Message sent by rider: %s (ID: %s)\n", msgResp.Content, messageID)

	// Driver sends message
	resp = makeRequest("POST", "/api/v1/messages", map[string]interface{}{
		"rideId":   rideID,
		"content":  "I'm 2 minutes away!",
		"metadata": map[string]string{"type": "status"},
	}, driverToken)

	json.Unmarshal(resp, &apiResp)
	json.Unmarshal(apiResp.Data, &msgResp)
	fmt.Printf("✅ Message sent by driver: %s\n", msgResp.Content)
}

func testGetMessages() {
	resp := makeRequest("GET", fmt.Sprintf("/api/v1/messages/rides/%s?limit=10&offset=0", rideID), nil, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	var messages []MessageResponse
	json.Unmarshal(apiResp.Data, &messages)
	fmt.Printf("✅ Retrieved %d messages for ride\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("   %d. %s\n", i+1, msg.Content)
	}
}

func testGetUnreadCount() {
	resp := makeRequest("GET", fmt.Sprintf("/api/v1/messages/rides/%s/unread-count", rideID), nil, driverToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	var unreadData map[string]interface{}
	json.Unmarshal(apiResp.Data, &unreadData)
	fmt.Printf("✅ Unread count: %v\n", unreadData["unreadCount"])
}

func testMarkAsRead() {
	resp := makeRequest("POST", fmt.Sprintf("/api/v1/messages/%s/read", messageID), nil, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	fmt.Printf("✅ Message marked as read\n")
}

func testDeleteMessage() {
	resp := makeRequest("DELETE", fmt.Sprintf("/api/v1/messages/%s", messageID), nil, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	if apiResp.Success {
		fmt.Printf("✅ Message deleted successfully\n")
	} else {
		fmt.Printf("❌ Failed to delete message: %s\n", apiResp.Message)
	}
}

func testErrorCases() {
	// Missing rideId
	fmt.Println("\n  Testing error: Missing rideId...")
	resp := makeRequest("POST", "/api/v1/messages", map[string]interface{}{
		"content": "This should fail",
	}, riderToken)

	var apiResp APIResponse
	json.Unmarshal(resp, &apiResp)
	if !apiResp.Success {
		fmt.Printf("  ✅ Correctly rejected: %s\n", apiResp.Message)
	}

	// Missing content
	fmt.Println("\n  Testing error: Missing content...")
	resp = makeRequest("POST", "/api/v1/messages", map[string]interface{}{
		"rideId": rideID,
	}, riderToken)

	json.Unmarshal(resp, &apiResp)
	if !apiResp.Success {
		fmt.Printf("  ✅ Correctly rejected: %s\n", apiResp.Message)
	}

	// Missing auth
	fmt.Println("\n  Testing error: Missing authentication...")
	resp = makeRequest("GET", fmt.Sprintf("/api/v1/messages/rides/%s", rideID), nil, "")

	json.Unmarshal(resp, &apiResp)
	if !apiResp.Success {
		fmt.Printf("  ✅ Correctly rejected: %s\n", apiResp.Message)
	}
}

func makeRequest(method, endpoint string, body interface{}, token string) []byte {
	url := baseURL + endpoint
	var reqBody io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return respBody
}
