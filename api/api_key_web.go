package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"

	// 为了方便演示，这里假设 tokens 文件路径是固定的
	"os"
)

// 假设你的 key 文件路径

const tokensFilePath = "keys/tokens" // 或者根据你的需求配置

// TokensUploadRequest 用于接收前端上传的 tokens 数据
type TokensUploadRequest struct {
	Tokens []string `json:"tokens"`
}

// CountResponse 用于返回 Tokens 数量
type CountResponse struct {
	Count int `json:"count"`
}

// ErrorTokensResponse 用于返回错误 Tokens 列表
type ErrorTokensResponse struct {
	Errors []string `json:"errors"`
}

// HandleUploadTokens 处理前端上传 Tokens 的请求
func HandleUploadTokens(w http.ResponseWriter, r *http.Request) {
	// 假设这里需要一些认证或权限验证（根据实际情况添加）
	// if !isAuthenticated(r) {
	//     http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//     return
	// }

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req TokensUploadRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 将接收到的 tokens 写入文件
	// 注意：这里直接覆盖了原有文件内容。如果你想追加或合并，需要更复杂的逻辑。
	// 同时需要考虑并发写入的问题，可能需要一个文件锁。

	keys := viper.GetString("Nkey.path")

	file, err := os.OpenFile(keys, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open tokens file for writing: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, token := range req.Tokens {
		_, err := writer.WriteString(token + "\n")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write token to file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to flush tokens file: %v", err), http.StatusInternalServerError)
		return
	}

	// 清空当前的 keyStatusMap，以便重新加载和标记密钥
	keyStatusMutex.Lock()
	// 清空 map
	for k := range keyStatusMap {
		delete(keyStatusMap, k)
	}
	// 在实际应用中，这里应该重新加载文件并更新 keyStatusMap
	// 但为了简化，这里只是清空 map，下一次 GetRandomKey 会重新加载文件

	keyStatusMutex.Unlock()

	// 通知等待者（如果有）密钥列表可能已更新
	keyStatusCond.Broadcast()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tokens uploaded successfully"))
}

// HandleGetAvailableTokensCount 处理获取可用 Tokens 数量的请求
func HandleGetAvailableTokensCount(w http.ResponseWriter, r *http.Request) {
	// 假设这里不需要认证

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 在访问 keyStatusMap 之前锁定 mutex
	keyStatusMutex.Lock()
	defer keyStatusMutex.Unlock()

	// 从文件中读取所有 key
	keys := viper.GetString("Nkey.path")

	allKeys, fileErr := getAllKeysFromFile(keys) // Reuse your existing function
	if fileErr != nil {
		// 文件读取失败，返回 0 或错误
		http.Error(w, fmt.Sprintf("Failed to read tokens file: %v", fileErr), http.StatusInternalServerError)
		// unlock is deferred
		return
	}

	totalKeys := len(allKeys)
	lockedKeysCount := 0
	// 统计被锁定的 key 数量
	for _, key := range allKeys {
		if keyStatusMap[keyStatusKey(key)] {
			lockedKeysCount++
		}
	}

	availableKeysCount := totalKeys - lockedKeysCount

	resp := CountResponse{Count: availableKeysCount}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleClearTokens 处理清空 Tokens 的请求（需要你实现清空逻辑）
func HandleClearTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 清空文件内容
	keys := viper.GetString("Nkey.path")

	file, err := os.OpenFile(keys, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to clear tokens file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// 文件已被截断，内容已经被清空

	// 清空内存中的 keyStatusMap
	keyStatusMutex.Lock()
	for k := range keyStatusMap {
		delete(keyStatusMap, k)
	}
	keyStatusMutex.Unlock()

	// 通知等待者（如果有）
	keyStatusCond.Broadcast()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tokens cleared successfully"))
}

// HandleGetErrorTokens 处理获取错误 Tokens 的请求（需要你实现错误 Tokens 的存储和管理）
func HandleGetErrorTokens(w http.ResponseWriter, r *http.Request) {
	// 获取错误 Token 文件的路径
	// 假设 viper 配置中保存了文件路径，例如 "Nkey.path_err"
	// 您需要确保 viper 已经初始化并加载了配置
	errorTokensFilePath := viper.GetString("Nkey.path_err")
	if errorTokensFilePath == "" {
		// 如果 viper 配置没有找到路径，或者路径为空，返回错误
		http.Error(w, "Error tokens file path not configured", http.StatusInternalServerError)
		return
	}

	// 从文件中读取所有 Token
	// 注意：这里假设 `keys/tokens_err` 文件中的所有 Token 都是“错误”的
	// 如果您有额外的逻辑来判断哪些是错误 Token，您需要修改 getAllKeysFromFile 或在读取后进行过滤
	errorTokens, fileErr := getAllKeysFromFile(errorTokensFilePath)
	if fileErr != nil {
		// 文件读取失败，返回错误给客户端
		http.Error(w, fmt.Sprintf("Failed to read error tokens file: %v", fileErr), http.StatusInternalServerError)
		return
	}

	// 构建要返回给前端的 JSON 结构
	resp := struct {
		Errors []string `json:"errors"` // 匹配前端预期的 "errors" 键
	}{
		Errors: errorTokens,
	}

	w.Header().Set("Content-Type", "application/json")
	// 设置响应头，允许跨域访问（如果前端和后端不在同一个域）
	// w.Header().Set("Access-Control-Allow-Origin", "*") // 仅用于开发测试，生产环境请指定具体的允许来源
	json.NewEncoder(w).Encode(resp)
}

// 在你的主应用中注册这些处理函数
/*
func main() {
    // ... 其他设置 ...
    http.HandleFunc("/tokens/upload", HandleUploadTokens)
    http.HandleFunc("/tokens/count", HandleGetAvailableTokensCount)
    http.HandleFunc("/tokens/errors", HandleGetErrorTokens) // 如果你实现了这个接口
    http.HandleFunc("/tokens", HandleClearTokens) // 使用 DELETE 方法清空
    // ... 启动服务器 ...
}
*/
