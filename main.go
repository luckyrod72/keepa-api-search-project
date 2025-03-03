package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 定义处理 HTTP 请求的函数
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/hello", handleHello)

	// 获取端口号，Cloud Run 会通过 PORT 环境变量提供端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
		log.Printf("使用默认端口 %s", port)
	}

	// 启动 HTTP 服务器
	log.Printf("开始监听端口 %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP 服务器启动失败: %v", err)
	}
}

// 根路径的处理函数
func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到请求: %s %s", r.Method, r.URL.Path)

	// 如果不是根路径，返回 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "欢迎访问 Golang HTTP 服务! 尝试访问 /hello 路径")
}

// /hello 路径的处理函数
func handleHello(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到请求: %s %s", r.Method, r.URL.Path)
	fmt.Fprintf(w, "你好，世界! 这是一个运行在 Google Cloud Run 上的 Go 服务")
}
