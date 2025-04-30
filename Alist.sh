#!/bin/bash

# 检查参数数量
if [ $# -ne 6 ]; then
    echo "Usage: $0 --username <username> --password <password> <local-file> <alist-url>"
    exit 1
fi

# 解析输入参数
USERNAME=$2
PASSWORD=$4
LOCAL_FILE=$5
REMOTE_PATH=$6

# 从 URL 中提取 alist app 的 base URL
BASE_URL=$(echo $REMOTE_PATH | sed -E 's|(https?://[^/]+).*|\1|')

# 提取 URL 路径部分（去除 BASE_URL）
REMOTE_FILE_PATH=$(echo $REMOTE_PATH | sed "s|$BASE_URL||")

# 获取 token
RESPONSE=$(curl -s --location --request POST "$BASE_URL/api/auth/login" \
-H 'Content-Type: application/json' \
-d "{ \"username\": \"$USERNAME\", \"password\": \"$PASSWORD\" }") > /dev/null 2>&1

# 提取 token
TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

if [ -z "$TOKEN" ]; then
    echo "Error: Failed to get token"
    exit 1
fi

# 上传文件
UPLOAD_URL="$BASE_URL/api/fs/put"
curl -s -# -T "$LOCAL_FILE" "$UPLOAD_URL" \
  -H "Authorization: $TOKEN" \
  -H "File-Path: $REMOTE_FILE_PATH/$LOCAL_FILE" > /dev/null 2>&1

# 提示用户上传完成
if [ $? -ne 0 ]; then
    exit 1
else
    rm -rf ./$LOCAL_FILE
fi


# 获取文件路径
response=$(curl -s -X POST "$BASE_URL/api/fs/get" \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d "{\"path\":\"${REMOTE_FILE_PATH}/${LOCAL_FILE}\",\"password\":\"\"}" | grep -o '"raw_url":"[^"]*"' | sed 's/"raw_url":"//' | sed 's/"//' )

echo "$response"

## 获取文件路径
#response=$(curl -s -X POST "$BASE_URL/api/fs/get" \
#  -H "Content-Type: application/json" \
#  -H "Authorization: $TOKEN" \
#  -d "{\"path\":\"${REMOTE_FILE_PATH}/${LOCAL_FILE}\",\"password\":\"\"}") > /dev/null 2>&1
#
## 检查 curl 命令是否成功 (非零退出状态码通常表示网络或连接问题)
#if [ $? -ne 0 ]; then
#  echo "Error: curl command failed."
#  exit 1
#fi
#
## 使用 jq 解析 JSON 响应并提取 raw_url
#raw_url=$(echo "$response" | jq -r '.data.raw_url')
#error_message=$(echo "$response" | jq -r '.message')
#response_code=$(echo "$response" | jq -r '.code') # 获取响应代码
#
## 检查 AList API 响应的状态码和raw_url
#if [ "$response_code" -eq 200 ] && [ -n "$raw_url" ] && [ "$raw_url" != "null" ]; then
#  echo "$raw_url"
#else
#  echo "原始响应: $response"
#  exit 1
#fi