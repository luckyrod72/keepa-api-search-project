package main

import (
	"cloud.google.com/go/firestore"
	memorystore "cloud.google.com/go/redis/apiv1"
	"cloud.google.com/go/redis/apiv1/redispb"
	"context"
	"crypto/tls"
	"crypto/x509"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

// Add these constants for Redis
const (
	// ... existing constants
	RedisKeyPrefix = "keepa:product:"
	RedisTTL       = 24 * time.Hour
)

// Add Redis client as a global variable
var redisClient *redis.Client

// Add Firestore client as a global variable
var firestoreClient *firestore.Client

func init() {
	// Configure Redis options
	ctx := context.Background()

	adminClient, _ := memorystore.NewCloudRedisClient(ctx)

	defer adminClient.Close()

	// Initialize Redis client
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	projectID := getEnv("PROJECT_ID", "")
	location := getEnv("REGION", "")
	instanceID := getEnv("INSTANCE_ID", "")

	req := &redispb.GetInstanceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectID, location, instanceID),
	}

	instance, _ := adminClient.GetInstance(ctx, req)

	// Load CA cert
	caCerts := instance.GetServerCaCerts()

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCerts[0].Cert))

	redisOptions := &redis.Options{
		Addr:         redisAddr,
		Password:     redisPassword,
		DB:           redisDB,
		PoolSize:     10,              // 连接池大小
		MinIdleConns: 2,               // 最小空闲连接数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 获取连接的超时时间
		TLSConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	redisClient = redis.NewClient(redisOptions)

	// Test Redis connection

	_, _ = redisClient.Ping(ctx).Result()

	// 启动健康检查 goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每 30 秒检查一次
		defer ticker.Stop()
		for range ticker.C {
			_ = redisClient.Ping(ctx).Err()
		}
	}()

	// Initialize Firestore client
	conf := &firebase.Config{ProjectID: projectID}
	app, _ := firebase.NewApp(ctx, conf)

	firestoreClient, _ = app.Firestore(ctx)

}

func main() {
	// Initialize Keepa client
	client := NewKeepaClient()

	// Initialize Gin router
	r := gin.Default()

	// Endpoint: Trigger Product Finder and Product Request
	r.POST("/keepa", client.handleFetchProducts)

	port := getEnv("PORT", "8080")

	// Start HTTP server
	client.Logger.Printf("Starting server on port %s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		client.Logger.Fatalf("Failed to start server: %v", err)
	}
}
