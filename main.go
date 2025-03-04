package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// KeepaResponse represents the structure of the Keepa API response
type KeepaResponse struct {
	Timestamp          int64    `json:"timestamp"`
	TokensLeft         int      `json:"tokensLeft"`
	RefillIn           int      `json:"refillIn"`
	RefillRate         int      `json:"refillRate"`
	TokenFlowReduction float64  `json:"tokenFlowReduction"`
	TokensConsumed     int      `json:"tokensConsumed"`
	ProcessingTimeInMs int      `json:"processingTimeInMs"`
	AsinList           []string `json:"asinList"`
	TotalResults       int      `json:"totalResults"`
}

// ProductDetailResponse represents the structure of the product detail API response
type ProductDetailResponse struct {
	ASIN    string      `json:"asin"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CombinedResponse represents the combined response with Keepa data and product details
type CombinedResponse struct {
	KeepaData      KeepaResponse           `json:"keepaData"`
	ProductDetails []ProductDetailResponse `json:"productDetails"`
}

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Keepa API proxy service")

	// Create a default gin router with logging
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Define a route to handle Keepa API requests
	r.POST("/keepa", handleKeepaQuery)

	// Run the server on the specified port
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}

// Handle Keepa API query requests
func handleKeepaQuery(c *gin.Context) {
	requestID := generateRequestID()
	startTime := time.Now()
	log.Printf("[%s] Received new request", requestID)

	// Get Keepa API URL and credentials from environment variables
	domain := getEnv("KEEPA_DOMAIN", "1")
	apiKey := getEnv("KEEPA_API_KEY", "rt7t1904up7638ddhboifgfksfedu7pap6gde8p5to6mtripoib3q4n1h3433rh4")
	url := fmt.Sprintf("https://api.keepa.com/query?domain=%s&key=%s", domain, apiKey)
	method := "POST"

	log.Printf("[%s] Using Keepa API endpoint: %s (domain: %s)", requestID, url, domain)

	// Parse JSON data from the request
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("[%s] Error parsing request JSON: %v", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request data: %v", err),
		})
		return
	}

	// Log request parameters (excluding sensitive data)
	logRequestParams(requestID, requestData)

	// Convert request data to JSON string
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("[%s] Error marshaling JSON data: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to process request data: %v", err),
		})
		return
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[%s] Error creating HTTP request: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create request: %v", err),
		})
		return
	}

	// Set request headers
	req.Header.Add("Content-Type", "application/json")
	log.Printf("[%s] Sending request to Keepa API", requestID)

	// Send request
	apiStartTime := time.Now()
	res, err := client.Do(req)
	apiDuration := time.Since(apiStartTime)

	if err != nil {
		log.Printf("[%s] Error sending request to Keepa API: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to send request: %v", err),
		})
		return
	}
	defer res.Body.Close()

	log.Printf("[%s] Received response from Keepa API - Status: %d, Duration: %v",
		requestID, res.StatusCode, apiDuration)

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[%s] Error reading response body: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read response: %v", err),
		})
		return
	}

	// Parse Keepa API response
	var keepaResponse KeepaResponse
	if err := json.Unmarshal(body, &keepaResponse); err != nil {
		log.Printf("[%s] Error parsing Keepa API response: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to parse Keepa API response: %v", err),
		})
		return
	}

	log.Printf("[%s] Successfully parsed Keepa response with %d ASINs", requestID, len(keepaResponse.AsinList))

	// Fetch product details for each ASIN
	productDetails := fetchProductDetails(requestID, keepaResponse.AsinList)

	// Combine Keepa response with product details
	combinedResponse := CombinedResponse{
		KeepaData:      keepaResponse,
		ProductDetails: productDetails,
	}

	// Return combined response to the client
	log.Printf("[%s] Returning combined response to client - Total request duration: %v",
		requestID, time.Since(startTime))
	c.JSON(http.StatusOK, combinedResponse)
}

// Fetch product details for a list of ASINs
func fetchProductDetails(requestID string, asinList []string) []ProductDetailResponse {
	productDetailURL := getEnv("PRODUCT_DETAIL_URL", "https://keepa-project-detail-937025550093.us-central1.run.app/product")
	log.Printf("[%s] Fetching product details for %d ASINs from %s", requestID, len(asinList), productDetailURL)

	// Create a client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second, // Longer timeout for multiple requests
	}

	// Use a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	// Use a mutex to protect concurrent writes to the results slice
	var mutex sync.Mutex
	// Prepare results slice
	results := make([]ProductDetailResponse, 0, len(asinList))

	// Limit concurrent requests
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	// Process each ASIN
	for _, asin := range asinList {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(asin string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			startTime := time.Now()
			log.Printf("[%s] Fetching product details for ASIN: %s", requestID, asin)

			// Create request URL with ASIN as query parameter
			url := fmt.Sprintf("%s?asin=%s", productDetailURL, asin)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("[%s] Error creating request for ASIN %s: %v", requestID, asin, err)
				mutex.Lock()
				results = append(results, ProductDetailResponse{
					ASIN:    asin,
					Success: false,
					Error:   fmt.Sprintf("Failed to create request: %v", err),
				})
				mutex.Unlock()
				return
			}

			// Send request
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("[%s] Error fetching product details for ASIN %s: %v", requestID, asin, err)
				mutex.Lock()
				results = append(results, ProductDetailResponse{
					ASIN:    asin,
					Success: false,
					Error:   fmt.Sprintf("Request failed: %v", err),
				})
				mutex.Unlock()
				return
			}
			defer resp.Body.Close()

			// Read response body
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[%s] Error reading response for ASIN %s: %v", requestID, asin, err)
				mutex.Lock()
				results = append(results, ProductDetailResponse{
					ASIN:    asin,
					Success: false,
					Error:   fmt.Sprintf("Failed to read response: %v", err),
				})
				mutex.Unlock()
				return
			}

			// Parse response
			var productDetail map[string]interface{}
			if err := json.Unmarshal(respBody, &productDetail); err != nil {
				log.Printf("[%s] Error parsing response for ASIN %s: %v", requestID, asin, err)
				mutex.Lock()
				results = append(results, ProductDetailResponse{
					ASIN:    asin,
					Success: false,
					Error:   fmt.Sprintf("Failed to parse response: %v", err),
				})
				mutex.Unlock()
				return
			}

			// Add result to the results slice
			mutex.Lock()
			results = append(results, ProductDetailResponse{
				ASIN:    asin,
				Success: true,
				Data:    productDetail,
			})
			mutex.Unlock()

			log.Printf("[%s] Successfully fetched product details for ASIN %s in %v",
				requestID, asin, time.Since(startTime))
		}(asin)
	}

	// Wait for all requests to complete
	wg.Wait()
	log.Printf("[%s] Completed fetching product details for all ASINs", requestID)

	return results
}

// Get environment variable or return default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Generate a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Log request parameters without sensitive information
func logRequestParams(requestID string, params map[string]interface{}) {
	// Create a copy of params to avoid modifying the original
	logParams := make(map[string]interface{})
	for k, v := range params {
		// Skip sensitive fields or replace with placeholder
		if k == "key" || k == "apiKey" || k == "token" || k == "password" || k == "secret" {
			logParams[k] = "[REDACTED]"
		} else {
			logParams[k] = v
		}
	}

	// Log the parameters
	paramJSON, err := json.Marshal(logParams)
	if err != nil {
		log.Printf("[%s] Request parameters: [failed to serialize]", requestID)
		return
	}

	log.Printf("[%s] Request parameters: %s", requestID, string(paramJSON))
}
