package main

import (
	"log"
	"time"
)

// Task represents the state of a task
type Task struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"` // "pending", "completed", "failed"
	ASINs      []string   `json:"asins,omitempty"`
	Products   []string   `json:"products,omitempty"` // Stores historical data for each ASIN
	Error      string     `json:"error,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Progress   int        `json:"progress"` // Number of ASINs processed so far
	Total      int        `json:"total"`    // Total number of ASINs to process
}

// KeepaClient represents a Keepa API client
type KeepaClient struct {
	TokensLeft      int
	RefillRate      float64
	SafetyThreshold int
	MaxRetries      int
	Logger          *log.Logger
	LastTimestamp   int64 // Last request timestamp for precise token recovery calculation
}

type APIResponse struct {
	Timestamp          int64          `json:"timestamp"`
	TokensLeft         int            `json:"tokensLeft"`
	RefillIn           int            `json:"refillIn"`
	RefillRate         int            `json:"refillRate"`
	TokenFlowReduction float64        `json:"tokenFlowReduction"`
	TokensConsumed     int            `json:"tokensConsumed"`
	ProcessingTimeInMs int            `json:"processingTimeInMs"`
	AsinList           []string       `json:"asinList"`
	Products           []KeepaProduct `json:"products"`
	TotalResults       int            `json:"totalResults"`
}

// Offer represents a single marketplace offer
type Offer struct {
	LastSeen         int         `json:"lastSeen"`
	SellerID         string      `json:"sellerId"`
	OfferCSV         []int       `json:"offerCSV"`
	Condition        int         `json:"condition"`
	ConditionComment interface{} `json:"conditionComment"`
	IsPrime          bool        `json:"isPrime"`
	IsMAP            bool        `json:"isMAP"`
	IsShippable      bool        `json:"isShippable"`
	IsAddonItem      bool        `json:"isAddonItem"`
	IsPreorder       bool        `json:"isPreorder"`
	IsWarehouseDeal  bool        `json:"isWarehouseDeal"`
	IsScam           bool        `json:"isScam"`
	IsAmazon         bool        `json:"isAmazon"`
	IsPrimeExcl      bool        `json:"isPrimeExcl"`
	OfferID          int         `json:"offerId"`
	StockCSV         []int       `json:"stockCSV"`
	IsFBA            bool        `json:"isFBA"`
	ShipsFromChina   bool        `json:"shipsFromChina"`
	StockLimit       []int       `json:"stockLimit"`
	MinOrderQty      int         `json:"minOrderQty"`
	CouponHistory    []int       `json:"couponHistory"`
}

// FBAFees represents Amazon FBA fees
type FBAFees struct {
	LastUpdate     int `json:"lastUpdate"`
	PickAndPackFee int `json:"pickAndPackFee"`
}

// Variation represents product variations
type Variation struct {
	Asin       string      `json:"asin"`
	Attributes []Attribute `json:"attributes"`
}

// Attribute represents a variation attribute
type Attribute struct {
	Dimension string `json:"dimension"`
	Value     string `json:"value"`
}

// UnitCount represents product unit information
type UnitCount struct {
	UnitValue float64 `json:"unitValue"`
	UnitType  string  `json:"unitType"`
}

// CategoryTreeItem represents an item in the category hierarchy
type CategoryTreeItem struct {
	CatID int    `json:"catId"`
	Name  string `json:"name"`
}

// BuyBoxSellerStats represents statistics for a seller in the buy box
type BuyBoxSellerStats struct {
	PercentageWon     float64 `json:"percentageWon"`
	AvgPrice          int     `json:"avgPrice"`
	AvgNewOfferCount  int     `json:"avgNewOfferCount"`
	AvgUsedOfferCount int     `json:"avgUsedOfferCount"`
	IsFBA             bool    `json:"isFBA"`
	LastSeen          int     `json:"lastSeen"`
	Condition         int     `json:"condition,omitempty"` // Only used in BuyBoxUsedStats
}

// ProductStats represents statistics for the product
type ProductStats struct {
	Current                        []int                        `json:"current"`
	Avg                            []int                        `json:"avg"`
	Avg30                          []int                        `json:"avg30"`
	Avg90                          []int                        `json:"avg90"`
	Avg180                         []int                        `json:"avg180"`
	Avg365                         []int                        `json:"avg365"`
	AtIntervalStart                []int                        `json:"atIntervalStart"`
	Min                            []interface{}                `json:"min"`
	Max                            []interface{}                `json:"max"`
	MinInInterval                  []interface{}                `json:"minInInterval"`
	MaxInInterval                  []interface{}                `json:"maxInInterval"`
	IsLowest                       []bool                       `json:"isLowest"`
	IsLowest90                     []bool                       `json:"isLowest90"`
	OutOfStockPercentageInInterval []int                        `json:"outOfStockPercentageInInterval"`
	OutOfStockPercentage365        []int                        `json:"outOfStockPercentage365"`
	OutOfStockPercentage180        []int                        `json:"outOfStockPercentage180"`
	OutOfStockPercentage90         []int                        `json:"outOfStockPercentage90"`
	OutOfStockPercentage30         []int                        `json:"outOfStockPercentage30"`
	OutOfStockCountAmazon30        int                          `json:"outOfStockCountAmazon30"`
	OutOfStockCountAmazon90        int                          `json:"outOfStockCountAmazon90"`
	DeltaPercent90MonthlySold      int                          `json:"deltaPercent90_monthlySold"`
	StockPerCondition3RdFBA        []int                        `json:"stockPerCondition3rdFBA"`
	StockPerConditionFBM           []int                        `json:"stockPerConditionFBM"`
	RetrievedOfferCount            int                          `json:"retrievedOfferCount"`
	TotalOfferCount                int                          `json:"totalOfferCount"`
	TradeInPrice                   int                          `json:"tradeInPrice"`
	LastOffersUpdate               int                          `json:"lastOffersUpdate"`
	IsAddonItem                    bool                         `json:"isAddonItem"`
	LightningDealInfo              interface{}                  `json:"lightningDealInfo"`
	SellerIdsLowestFBA             []string                     `json:"sellerIdsLowestFBA"`
	SellerIdsLowestFBM             []string                     `json:"sellerIdsLowestFBM"`
	OfferCountFBA                  int                          `json:"offerCountFBA"`
	OfferCountFBM                  int                          `json:"offerCountFBM"`
	SalesRankDrops30               int                          `json:"salesRankDrops30"`
	SalesRankDrops90               int                          `json:"salesRankDrops90"`
	SalesRankDrops180              int                          `json:"salesRankDrops180"`
	SalesRankDrops365              int                          `json:"salesRankDrops365"`
	BuyBoxPrice                    int                          `json:"buyBoxPrice"`
	BuyBoxShipping                 int                          `json:"buyBoxShipping"`
	BuyBoxIsUnqualified            bool                         `json:"buyBoxIsUnqualified"`
	BuyBoxIsShippable              bool                         `json:"buyBoxIsShippable"`
	BuyBoxIsPreorder               bool                         `json:"buyBoxIsPreorder"`
	BuyBoxIsFBA                    bool                         `json:"buyBoxIsFBA"`
	BuyBoxIsAmazon                 bool                         `json:"buyBoxIsAmazon"`
	BuyBoxIsMAP                    bool                         `json:"buyBoxIsMAP"`
	BuyBoxIsUsed                   bool                         `json:"buyBoxIsUsed"`
	BuyBoxIsBackorder              bool                         `json:"buyBoxIsBackorder"`
	BuyBoxIsPrimeExclusive         bool                         `json:"buyBoxIsPrimeExclusive"`
	BuyBoxIsFreeShippingEligible   bool                         `json:"buyBoxIsFreeShippingEligible"`
	BuyBoxIsPrimePantry            bool                         `json:"buyBoxIsPrimePantry"`
	BuyBoxIsPrimeEligible          bool                         `json:"buyBoxIsPrimeEligible"`
	BuyBoxMinOrderQuantity         int                          `json:"buyBoxMinOrderQuantity"`
	BuyBoxMaxOrderQuantity         int                          `json:"buyBoxMaxOrderQuantity"`
	BuyBoxCondition                int                          `json:"buyBoxCondition"`
	LastBuyBoxUpdate               int                          `json:"lastBuyBoxUpdate"`
	BuyBoxAvailabilityMessage      interface{}                  `json:"buyBoxAvailabilityMessage"`
	BuyBoxShippingCountry          interface{}                  `json:"buyBoxShippingCountry"`
	BuyBoxSellerID                 string                       `json:"buyBoxSellerId"`
	BuyBoxIsWarehouseDeal          bool                         `json:"buyBoxIsWarehouseDeal"`
	BuyBoxStats                    map[string]BuyBoxSellerStats `json:"buyBoxStats"`
	BuyBoxUsedStats                map[string]BuyBoxSellerStats `json:"buyBoxUsedStats"`
}

// AutoGenerated is the main product data structure
type KeepaProduct struct {
	Csv                             []interface{}      `json:"csv"`
	Categories                      []int64            `json:"categories"`
	ImagesCSV                       string             `json:"imagesCSV"`
	Manufacturer                    string             `json:"manufacturer"`
	Title                           string             `json:"title"`
	LastUpdate                      int                `json:"lastUpdate"`
	LastPriceChange                 int                `json:"lastPriceChange"`
	RootCategory                    int                `json:"rootCategory"`
	ProductType                     int                `json:"productType"`
	ParentAsin                      string             `json:"parentAsin"`
	VariationCSV                    string             `json:"variationCSV"`
	Asin                            string             `json:"asin"`
	DomainID                        int                `json:"domainId"`
	Type                            string             `json:"type"`
	HasReviews                      bool               `json:"hasReviews"`
	TrackingSince                   int                `json:"trackingSince"`
	Brand                           string             `json:"brand"`
	ProductGroup                    string             `json:"productGroup"`
	PartNumber                      string             `json:"partNumber"`
	Model                           string             `json:"model"`
	Color                           string             `json:"color"`
	Size                            string             `json:"size"`
	Edition                         interface{}        `json:"edition"`
	Format                          interface{}        `json:"format"`
	PackageHeight                   int                `json:"packageHeight"`
	PackageLength                   int                `json:"packageLength"`
	PackageWidth                    int                `json:"packageWidth"`
	PackageWeight                   int                `json:"packageWeight"`
	PackageQuantity                 int                `json:"packageQuantity"`
	IsAdultProduct                  bool               `json:"isAdultProduct"`
	IsEligibleForTradeIn            bool               `json:"isEligibleForTradeIn"`
	IsEligibleForSuperSaverShipping bool               `json:"isEligibleForSuperSaverShipping"`
	Offers                          []Offer            `json:"offers"`
	BuyBoxSellerIDHistory           []string           `json:"buyBoxSellerIdHistory"`
	IsRedirectASIN                  bool               `json:"isRedirectASIN"`
	IsSNS                           bool               `json:"isSNS"`
	Author                          interface{}        `json:"author"`
	Binding                         string             `json:"binding"`
	NumberOfItems                   int                `json:"numberOfItems"`
	NumberOfPages                   int                `json:"numberOfPages"`
	PublicationDate                 int                `json:"publicationDate"`
	ReleaseDate                     int                `json:"releaseDate"`
	Languages                       interface{}        `json:"languages"`
	LastRatingUpdate                int                `json:"lastRatingUpdate"`
	EbayListingIds                  interface{}        `json:"ebayListingIds"`
	LastEbayUpdate                  int                `json:"lastEbayUpdate"`
	EanList                         []string           `json:"eanList"`
	UpcList                         []string           `json:"upcList"`
	LiveOffersOrder                 []int              `json:"liveOffersOrder"`
	FrequentlyBoughtTogether        []string           `json:"frequentlyBoughtTogether"`
	Features                        []string           `json:"features"`
	Description                     string             `json:"description"`
	Promotions                      interface{}        `json:"promotions"`
	NewPriceIsMAP                   bool               `json:"newPriceIsMAP"`
	Coupon                          interface{}        `json:"coupon"`
	AvailabilityAmazon              int                `json:"availabilityAmazon"`
	ListedSince                     int                `json:"listedSince"`
	FbaFees                         FBAFees            `json:"fbaFees"`
	Variations                      []Variation        `json:"variations"`
	ItemHeight                      int                `json:"itemHeight"`
	ItemLength                      int                `json:"itemLength"`
	ItemWidth                       int                `json:"itemWidth"`
	ItemWeight                      int                `json:"itemWeight"`
	SalesRankReference              int                `json:"salesRankReference"`
	SalesRanks                      map[string][]int   `json:"salesRanks"`
	SalesRankReferenceHistory       []int              `json:"salesRankReferenceHistory"`
	Launchpad                       bool               `json:"launchpad"`
	IsB2B                           bool               `json:"isB2B"`
	LastStockUpdate                 int                `json:"lastStockUpdate"`
	BuyBoxUsedHistory               []string           `json:"buyBoxUsedHistory"`
	LastSoldUpdate                  int                `json:"lastSoldUpdate"`
	MonthlySold                     int                `json:"monthlySold"`
	MonthlySoldHistory              []int              `json:"monthlySoldHistory"`
	BuyBoxEligibleOfferCounts       []int              `json:"buyBoxEligibleOfferCounts"`
	CompetitivePriceThreshold       int                `json:"competitivePriceThreshold"`
	ParentAsinHistory               []string           `json:"parentAsinHistory"`
	IsHeatSensitive                 bool               `json:"isHeatSensitive"`
	ReturnRate                      int                `json:"returnRate"`
	URLSlug                         string             `json:"urlSlug"`
	UnitCount                       UnitCount          `json:"unitCount"`
	ItemTypeKeyword                 string             `json:"itemTypeKeyword"`
	RecommendedUsesForProduct       string             `json:"recommendedUsesForProduct"`
	Style                           string             `json:"style"`
	IncludedComponents              string             `json:"includedComponents"`
	Material                        string             `json:"material"`
	BrandStoreName                  string             `json:"brandStoreName"`
	BrandStoreURL                   string             `json:"brandStoreUrl"`
	Stats                           ProductStats       `json:"stats"`
	OffersSuccessful                bool               `json:"offersSuccessful"`
	G                               int                `json:"g"`
	CategoryTree                    []CategoryTreeItem `json:"categoryTree"`
	ParentTitle                     string             `json:"parentTitle"`
	BrandStoreURLName               string             `json:"brandStoreUrlName"`
	ReferralFeePercent              int                `json:"referralFeePercent"`
	ReferralFeePercentage           float64            `json:"referralFeePercentage"`
}

// Create simplified response with only the needed fields
type SimplifiedOffer struct {
	SellerID  string         `json:"sellerId"`
	Condition int            `json:"condition"`
	IsPrime   bool           `json:"isPrime"`
	IsAmazon  bool           `json:"isAmazon"`
	IsFBA     bool           `json:"isFBA"`
	StockCSV  map[string]int `json:"stockCSV,omitempty"`
}

type SimplifiedProduct struct {
	Asin        string            `json:"asin"`
	Title       string            `json:"title"`
	Categories  []int64           `json:"categories"`
	Brand       string            `json:"brand"`
	BuyBoxPrice int               `json:"buyBoxPrice,omitempty"`
	SalesRanks  map[string]int    `json:"salesRanks,omitempty"`
	Offers      []SimplifiedOffer `json:"offers,omitempty"`
}

type SimplifiedResponse struct {
	Products []SimplifiedProduct `json:"products"`
}
