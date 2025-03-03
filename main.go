package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建一个默认的 gin 路由引擎
	r := gin.Default()

	// 定义一个处理 Keepa API 请求的路由
	r.GET("/keepa", handleKeepaQuery)

	// 运行服务器在 8080 端口
	r.Run(":8080") // 监听并在 0.0.0.0:8080 上服务
}

// 处理 Keepa API 请求的函数
func handleKeepaQuery(c *gin.Context) {
	// Keepa API 的 URL 和凭证
	url := "https://api.keepa.com/query?domain=1&key=rt7t1904up7638ddhboifgfksfedu7pap6gde8p5to6mtripoib3q4n1h3433rh4"
	method := "POST"

	// API 请求体
	payload := strings.NewReader(`{
		"page": 0,
		"perPage": 500,
		"rootCategory": 1055398,
		"salesRankReference": 1055398,
		"availabilityAmazon": 3,
		"hasReviews": true,
		"returnRate": 1,
		"buyBoxStatsAmazon": 30,
		"outOfStockCountAmazon90_gte": 5
	}`)

	// 创建 HTTP 客户端
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("创建请求失败: %v", err),
		})
		return
	}

	// 设置请求头
	req.Header.Add("Content-Type", "application/json")

	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("发送请求失败: %v", err),
		})
		return
	}
	defer res.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("读取响应失败: %v", err),
		})
		return
	}

	// 将 Keepa API 的响应原样返回给客户端
	c.Data(res.StatusCode, "application/json", body)
}
