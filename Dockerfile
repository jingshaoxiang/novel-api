# 使用更小的基础镜像
FROM alpine:latest

# 更新软件包列表并安装必要的依赖项
# 减少 RUN 指令的数量，合并为一条
RUN apk update && apk add --no-cache curl bash

# 设置工作目录
WORKDIR /root/

# 创建目录
# 合并为一条 RUN 指令
RUN mkdir ./keys

# 复制文件
# 尽量一次性复制多个文件，减少 COPY 指令的数量
COPY web ./web
COPY Alist.sh  ./
COPY Minio.sh  ./
COPY mc_aarch64  ./
COPY mc_x86_64  ./
COPY novel-x86  ./
COPY novel-arm  ./

# 设置执行权限
# 为可执行文件和脚本设置执行权限
RUN chmod +x Alist.sh Minio.sh mc_aarch64 mc_x86_64 novel-x86 novel-arm

# 暴露端口
EXPOSE 3388

# 执行命令
# 使用 exec 形式 ('[]') 运行命令，这是 Docker 推荐的方式
# 考虑根据架构选择运行哪个可执行文件（如果需要的话）
# 例如，可以在 CMD 中添加逻辑判断，或者创建单独的 Dockerfile
# 这里假设直接运行 novel-arm 是期望的行为
# 使用一个简单的 shell 脚本来判断架构并运行
CMD ["/bin/sh", "-c", "ARCH=$(uname -m); if [ \"$ARCH\" = \"aarch64\" ]; then ./novel-arm; elif [ \"$ARCH\" = \"x86_64\" ]; then ./novel-x86; else echo \"Unsupported architecture: $ARCH\"; exit 1; fi"]