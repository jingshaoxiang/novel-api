#!/bin/bash

# 判断主机类型
Arch=$(arch)

# 传参
MINIO_ALIAS=$1
MINIO_URL=$2
MINIO_AK=$3
MINIO_SK=$4
MINIO_FILE=$5
MINIO_NAME=$6

./mc_$Arch alias set $MINIO_ALIAS $MINIO_URL/ $MINIO_AK $MINIO_SK  &>/dev/null

# MinIO别名
MINIO_ALIAS=$MINIO_ALIAS
# 要上传的本地文件路径
LOCAL_FILE_PATH=./$MINIO_FILE
# 目标MinIO桶名称  指定文件夹的话，用/隔开即可，如下所示
BUCKET_NAME=$MINIO_NAME
# 上传到MinIO后的文件名（可选，如果不指定则与本地文件名相同）
MINIO_FILE_NAME=$(basename "$LOCAL_FILE_PATH")

# 检查本地文件是否存在
if [ ! -f "$LOCAL_FILE_PATH" ]; then
    exit 1
fi

# 上传文件至MinIO
./mc_$Arch cp "$LOCAL_FILE_PATH" "${MINIO_ALIAS}/${BUCKET_NAME}/${MINIO_FILE_NAME}"  &>/dev/null

# 检查上传是否成功
if [ $? -ne 0 ]; then
    exit 1
fi

# 设置下载权限
./mc_$Arch anonymous set download $MINIO_ALIAS/$MINIO_NAME/$MINIO_FILE &>/dev/null

# 检查上传是否成功
if [ $? -ne 0 ]; then
    exit 1
else
    rm -rf ./$MINIO_FILE
fi

# 返回下载链接
echo "$MINIO_URL/$MINIO_NAME/$MINIO_FILE"