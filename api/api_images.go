package api

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

func ImageURLToBase64(imageURL string) (string, error) {
	// 发送HTTP GET请求
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	// 读取响应体
	fileData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 将图像数据编码为Base64
	encodedString := base64.StdEncoding.EncodeToString(fileData)
	return encodedString, nil
}
