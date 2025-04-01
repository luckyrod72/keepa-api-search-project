package main

import "os"

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// calculateProductFinderTokens calculates token consumption for Product Finder
func calculateProductFinderTokens(numASINs int) int {
	baseCost := 10                     // Base cost
	extraCost := (numASINs + 99) / 100 // 1 extra token per 100 ASINs
	return baseCost + extraCost
}

// calculateProductRequestTokens calculates token consumption for Product Request (worst case)
func calculateProductRequestTokens(numASINs int) int {
	return numASINs * 2 // 2 tokens per ASIN (assuming refresh is needed)
}
