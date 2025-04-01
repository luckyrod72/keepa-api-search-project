package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// NewKeepaClient initializes a new Keepa client
func NewKeepaClient() *KeepaClient {
	// Initialize logger
	logger := log.New(os.Stdout, "KeepaClient: ", log.LstdFlags|log.Lshortfile)

	return &KeepaClient{
		TokensLeft:      300, // Initial token count
		RefillRate:      5.0, // 5 tokens per minute
		SafetyThreshold: 10,  // Safety threshold for tokens
		MaxRetries:      3,   // Maximum retry attempts
		Logger:          logger,
		LastTimestamp:   time.Now().UnixNano() / int64(time.Millisecond), // Initialize timestamp
	}
}

// updateTokens precisely calculates token recovery
func (client *KeepaClient) updateTokens(currentTimestamp int64) {
	// Calculate time difference (in milliseconds)
	timeDiffMs := float64(currentTimestamp - client.LastTimestamp)
	// Calculate recovered tokens (RefillRate tokens per minute)
	tokensRecovered := (timeDiffMs / 1000.0) * (client.RefillRate / 60.0)
	// Update token count
	client.TokensLeft += int(tokensRecovered)
	// Cap token count at 300
	if client.TokensLeft > 300 {
		client.TokensLeft = 300
	}
	// Update timestamp
	client.LastTimestamp = currentTimestamp
	client.Logger.Printf("Updated tokens: %d (recovered %.2f tokens)", client.TokensLeft, tokensRecovered)
}

// waitForTokens waits for token recovery if needed
func (client *KeepaClient) waitForTokens(requiredTokens int, refillIn int) {
	if client.TokensLeft >= requiredTokens {
		return
	}

	// Calculate wait time
	tokensNeeded := requiredTokens - client.TokensLeft
	secondsPerToken := 60.0 / client.RefillRate // Seconds per token
	waitSeconds := float64(tokensNeeded) * secondsPerToken

	// Use refillIn if provided
	if refillIn > 0 {
		waitSeconds = float64(refillIn) / 1000.0 // Convert to seconds
	}

	client.Logger.Printf("Tokens insufficient. Need %d, have %d. Waiting %.2f seconds...", requiredTokens, client.TokensLeft, waitSeconds)
	time.Sleep(time.Duration(waitSeconds * float64(time.Second)))

	// Simulate token recovery
	currentTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
	client.updateTokens(currentTimestamp)
}

// calculateDynamicBatchSize dynamically calculates batchSize based on current token count
func (client *KeepaClient) calculateDynamicBatchSize(maxBatchSize int) int {
	// Update token state
	currentTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
	client.updateTokens(currentTimestamp)

	// Calculate available tokens
	availableTokens := client.TokensLeft - client.SafetyThreshold
	if availableTokens <= 0 {
		return 1 // Process at least 1 ASIN
	}

	// Each ASIN consumes 2 tokens (worst case)
	maxASINs := availableTokens / 2
	if maxASINs > maxBatchSize {
		maxASINs = maxBatchSize
	}
	if maxASINs < 1 {
		maxASINs = 1
	}

	client.Logger.Printf("Calculated dynamic batchSize: %d (available tokens: %d)", maxASINs, availableTokens)
	return maxASINs
}

// doRequest is a generic request method with retry logic and exponential backoff
func (client *KeepaClient) doRequest(url string, requiredTokens int, method string, queryParam map[string]interface{}) (*APIResponse, error) {
	// Estimate token consumption and check if waiting is needed
	currentTimestamp := time.Now().UnixNano() / int64(time.Millisecond)
	client.updateTokens(currentTimestamp)

	if requiredTokens+client.SafetyThreshold > client.TokensLeft {
		client.waitForTokens(requiredTokens+client.SafetyThreshold, 0)
	}

	// Retry logic
	for retry := 0; retry <= client.MaxRetries; retry++ {
		client.Logger.Printf("Sending request to %s (retry %d/%d)", url, retry, client.MaxRetries)

		var resp *http.Response
		var err error

		if method == "GET" {
			resp, err = http.Get(url)
		}
		if method == "POST" {
			jsonData, err := json.Marshal(queryParam)
			if err != nil {
				log.Printf(" Error marshaling JSON data: %v", err)
			}
			resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		}

		if err != nil {
			client.Logger.Printf("HTTP request failed: %v", err)
			return nil, fmt.Errorf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode == http.StatusTooManyRequests { // 429
			client.Logger.Printf("Received 429 Too Many Requests, retry %d/%d", retry+1, client.MaxRetries)

			// Read response body to get refillIn and tokensLeft
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				client.Logger.Printf("Failed to read 429 response body: %v", err)
				return nil, fmt.Errorf("Failed to read 429 response body: %v", err)
			}

			var apiResp APIResponse
			if err := json.Unmarshal(body, &apiResp); err != nil {
				client.Logger.Printf("Failed to parse 429 response: %v", err)
				return nil, fmt.Errorf("Failed to parse 429 response: %v", err)
			}

			// Update token state
			client.TokensLeft = apiResp.TokensLeft
			client.LastTimestamp = apiResp.Timestamp
			client.Logger.Printf("429 Response: Tokens left: %d, Refill in: %d ms", client.TokensLeft, apiResp.RefillIn)

			// Return error if max retries reached
			if retry == client.MaxRetries {
				client.Logger.Printf("Max retries reached after 429 error")
				return nil, fmt.Errorf("Max retries reached after 429 error")
			}

			// Exponential backoff: wait time = base wait time + 2^retry seconds
			baseWaitSeconds := float64(apiResp.RefillIn) / 1000.0
			if baseWaitSeconds <= 0 {
				tokensNeeded := requiredTokens + client.SafetyThreshold - client.TokensLeft
				secondsPerToken := 60.0 / client.RefillRate
				baseWaitSeconds = float64(tokensNeeded) * secondsPerToken
			}
			retryWaitSeconds := baseWaitSeconds + math.Pow(2, float64(retry))
			client.Logger.Printf("Applying exponential backoff: Waiting %.2f seconds", retryWaitSeconds)

			time.Sleep(time.Duration(retryWaitSeconds * float64(time.Second)))
			// Update token state
			currentTimestamp = time.Now().UnixNano() / int64(time.Millisecond)
			client.updateTokens(currentTimestamp)
			continue
		}

		// Handle non-200 status codes
		if resp.StatusCode != http.StatusOK {
			client.Logger.Printf("Unexpected status code: %d", resp.StatusCode)
			return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
		}

		// Read response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			client.Logger.Printf("Failed to read response body: %v", err)
			return nil, fmt.Errorf("Failed to read response body: %v", err)
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			client.Logger.Printf("Failed to parse response: %v", err)
			return nil, fmt.Errorf("Failed to parse response: %v", err)
		}

		// Update token state
		client.TokensLeft = apiResp.TokensLeft
		client.LastTimestamp = apiResp.Timestamp
		return &apiResp, nil
	}

	return nil, fmt.Errorf("Unexpected error after retries")
}

// ProductFinder simulates a Product Finder API request
func (client *KeepaClient) ProductFinder(queryParam map[string]interface{}, pageSize int) ([]string, error) {
	// Estimate token consumption
	requiredTokens := calculateProductFinderTokens(pageSize)
	// Construct request URL
	domain := getEnv("KEEPA_DOMAIN", "1")
	apiKey := getEnv("KEEPA_API_KEY", "rt7t1904up7638ddhboifgfksfedu7pap6gde8p5to6mtripoib3q4n1h3433rh4")
	url := fmt.Sprintf("https://api.keepa.com/query?domain=%s&key=%s", domain, apiKey)

	// Send request
	apiResp, err := client.doRequest(url, requiredTokens, "POST", queryParam)
	if err != nil {
		return nil, err
	}

	client.Logger.Printf("Product Finder: Consumed %d tokens, %d tokens left, refill in %d ms", apiResp.TokensConsumed, client.TokensLeft, apiResp.RefillIn)
	return apiResp.AsinList, nil
}

// ProductRequest simulates a Product Request API request
func (client *KeepaClient) ProductRequest(asin string) (*SimplifiedResponse, error) {
	// Process only 1 ASIN at a time
	asins := []string{asin}
	// Estimate token consumption
	requiredTokens := calculateProductRequestTokens(len(asins))

	domain := getEnv("KEEPA_DOMAIN", "1")
	apiKey := getEnv("KEEPA_API_KEY", "rt7t1904up7638ddhboifgfksfedu7pap6gde8p5to6mtripoib3q4n1h3433rh4")
	stats := getEnv("KEEPA_STATS", "90")
	update := getEnv("KEEPA_UPDATE", "-1")
	history := getEnv("KEEPA_HISTORY", "1")
	days := getEnv("KEEPA_DAYS", "90")
	codeLimit := getEnv("KEEPA_CODE_LIMIT", "10")
	offers := getEnv("KEEPA_OFFERS", "20")
	onlyLiveOffers := getEnv("KEEPA_ONLY_LIVE_OFFERS", "1")
	rental := getEnv("KEEPA_RENTAL", "0")
	videos := getEnv("KEEPA_VIDEOS", "0")
	aplus := getEnv("KEEPA_APLUS", "0")
	rating := getEnv("KEEPA_RATING", "0")
	buybox := getEnv("KEEPA_BUYBOX", "1")
	stock := getEnv("KEEPA_STOCK", "1")

	// Construct request URL
	url := fmt.Sprintf("https://api.keepa.com/product?domain=%s&key=%s&asin=%s&stats=%s&update=%s&history=%s&days=%s&code-limit=%s&offers=%s&only-live-offers=%s&rental=%s&videos=%s&aplus=%s&rating=%s&buybox=%s&stock=%s",
		domain, apiKey, asin, stats, update, history, days, codeLimit, offers, onlyLiveOffers, rental, videos, aplus, rating, buybox, stock)

	// Send request
	apiResp, err := client.doRequest(url, requiredTokens, "GET", nil)
	if err != nil {
		return nil, err
	}

	client.Logger.Printf("Product Request: Consumed %d tokens, %d tokens left, refill in %d ms", apiResp.TokensConsumed, client.TokensLeft, apiResp.RefillIn)

	// Parse the Keepa API response
	simplifiedResponse := &SimplifiedResponse{Products: make([]SimplifiedProduct, 0)}
	for _, product := range apiResp.Products {
		rootCategory := strconv.Itoa(product.RootCategory)

		// Create sales ranks map with timestamp as key and rank as value
		salesRanks := make(map[string]int)
		if len(product.SalesRanks[rootCategory]) > 0 && len(product.SalesRanks[rootCategory])%2 == 0 {
			for i := 0; i < len(product.SalesRanks[rootCategory]); i += 2 {
				timestamp := time.UnixMilli(int64(product.SalesRanks[rootCategory][i]+21564000) * 60000)
				timestampStr := timestamp.Format(time.DateTime)
				salesRanks[timestampStr] = product.SalesRanks[rootCategory][i+1]
			}
		}

		simplifiedProduct := SimplifiedProduct{
			Asin:       product.Asin,
			Title:      product.Title,
			Categories: product.Categories,
			Brand:      product.Brand,
			SalesRanks: salesRanks,
		}

		// Add buyBoxPrice if available
		if product.Stats.BuyBoxPrice != 0 {
			simplifiedProduct.BuyBoxPrice = product.Stats.BuyBoxPrice
		}

		// Add simplified offers
		for _, offer := range product.Offers {
			simplifiedOffer := SimplifiedOffer{
				SellerID:  offer.SellerID,
				Condition: offer.Condition,
				IsPrime:   offer.IsPrime,
				IsAmazon:  offer.IsAmazon,
				IsFBA:     offer.IsFBA,
			}

			// Only include stockCSV if it's not empty
			if len(offer.StockCSV) > 0 && len(offer.StockCSV)%2 == 0 {
				stockCSV := make(map[string]int)
				for i := 0; i < len(offer.StockCSV); i += 2 {
					timestamp := time.UnixMilli(int64(offer.StockCSV[i]+21564000) * 60000)
					timestampStr := timestamp.Format(time.DateTime)
					stockCSV[timestampStr] = offer.StockCSV[i+1]
				}
				simplifiedOffer.StockCSV = stockCSV
			}

			simplifiedProduct.Offers = append(simplifiedProduct.Offers, simplifiedOffer)
		}

		simplifiedResponse.Products = append(simplifiedResponse.Products, simplifiedProduct)
	}
	return simplifiedResponse, nil
}

// createTask creates a new task

// handleFetchProducts handles Product Finder and Product Request requests
func (client *KeepaClient) handleFetchProducts(c *gin.Context) {

	taskID := generateTaskID()

	pageSize := 50

	// Get Keepa API URL and credentials from environment variables

	categoryList := getEnv("KEEPA_CATEGORY", "1055398;3760901;3760911;16310101;165796011;2619533011;3375251;228013;1064954;172282")
	categoryListArr := strings.Split(categoryList, ";")

	// Parse JSON data from the request
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request data: %v", err),
		})
		return
	}

	for _, category := range categoryListArr {
		requestData["rootCategory"] = category
		requestData["salesRankReference"] = category
		// Create task
		client.Logger.Printf("Created task %s for Fetch Products (pageSize: %d)", taskID, pageSize)

		// Step 1: Call Product Finder to get ASIN list
		asins, err := client.ProductFinder(requestData, pageSize)
		if err != nil {
			client.Logger.Printf("Task %s failed at Product Finder: %v", taskID, err)
			return
		}

		// Update task state
		client.Logger.Printf("Task %s: Retrieved %d ASINs from Product Finder", taskID, len(asins))

		// Step 2: Call Product Request for each ASIN individually
		for i, asin := range asins {
			var product *SimplifiedResponse

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Try to get data from Redis first
			if product, err = getProductFromRedis(ctx, asin); err == nil {
				if err = firestoreFunction(ctx, taskID, asin, product); err != nil {
					client.Logger.Printf("[RequestID: %s] Failed to save data to Firestore for ASIN %s: %v", taskID, asin, err)
					continue // Skip failed ASIN and continue with the next one
				}
				return
			}

			// Call Product Request for each ASIN individually and append the response to the allProducts slice
			product, err = client.ProductRequest(asin)
			if err != nil {
				client.Logger.Printf("Task %s: Failed to retrieve data for ASIN %s: %v", taskID, asin, err)
				continue // Skip failed ASIN and continue with the next one
			}

			// Save to Redis
			err = saveProductToRedis(ctx, asin, product)
			if err != nil {
				client.Logger.Printf("[RequestID: %s] Failed to save data to Redis for ASIN %s: %v", taskID, asin, err)
			}

			firestoreFunction(ctx, taskID, asin, product)

			client.Logger.Printf("Task %s: Retrieved data for ASIN %s (%d/%d)", taskID, asin, i+1, len(asins))
		}

		// Task completed
		client.Logger.Printf("Task %s completed: Processed %d ASINs", taskID, len(asins))
	}

	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": "pending"})
}

// Generate a unique Task ID for each request
func generateTaskID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
