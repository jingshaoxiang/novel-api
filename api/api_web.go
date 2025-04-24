package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 假设你的前端文件放在一个名为 "web" 的文件夹里
const staticFilesDir = "web" // 存放 index.html 和 script.js 的目录
// 处理 /web 路径的请求

func WebCheck(w http.ResponseWriter, r *http.Request) {
	// 构建请求的文件路径
	requestedPath := r.URL.Path
	// 去掉 /web 前缀，例如 /web/index.html -> /index.html
	if strings.HasPrefix(requestedPath, "/web/") {
		requestedPath = strings.TrimPrefix(requestedPath, "/web")
	} else if requestedPath == "/web" {
		// 如果只输入 /web，重定向到 /web/
		http.Redirect(w, r, "/web/", http.StatusMovedPermanently)
		return
	}

	// 如果请求路径是根路径 /，则默认返回 index.html
	if requestedPath == "/" {
		requestedPath = "/index.html"
	}

	// 构建完整的文件路径
	filePath := filepath.Join(staticFilesDir, requestedPath)

	// 重要的安全检查：防止路径遍历攻击
	// 确保拼接后的路径在 staticFilesDir 目录内
	absStaticFilesDir, err := filepath.Abs(staticFilesDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error getting absolute path for static directory: %v", err)
		return
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error getting absolute path for requested file: %v", err)
		return
	}

	// 检查请求的文件是否在允许的静态文件目录下
	if !strings.HasPrefix(absFilePath, absStaticFilesDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("Attempted path traversal: %s", requestedPath)
		return
	}

	// 检查文件是否存在
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error stating file %s: %v", filePath, err)
		return
	}

	// 使用 http.ServeFile 提供文件服务
	http.ServeFile(w, r, filePath)
}
