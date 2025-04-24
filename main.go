package main

import (
	"NoveAI3/api"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

func main() {
	//获取图片保存路径
	viper.SetConfigFile("config.yml")

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	} else {
		//fmt.Println("配置文件加载成功.......")
	}
	// 读取配置文件
	configFile, err := os.Open("config.yml")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	//Port := viper.GetString("start.port")
	Port := "3388"

	http.HandleFunc("/v1/chat/completions", api.Completions) // 修改了路由
	http.HandleFunc("/tokens/upload", api.HandleUploadTokens)
	http.HandleFunc("/tokens/count", api.HandleGetAvailableTokensCount)
	http.HandleFunc("/tokens", api.HandleClearTokens)           // 使用 DELETE 方法清空
	http.HandleFunc("/tokens/errors", api.HandleGetErrorTokens) // 如果你实现了这个接口
	http.HandleFunc("/web/", api.WebCheck)                      // 前端页面

	log.Println("Starting server on : ", Port)

	if err := http.ListenAndServe(":"+Port, nil); err != nil {
		log.Fatal(err)
	}
}
