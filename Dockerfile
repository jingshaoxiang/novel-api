# Step 2: Run the Go binary
FROM alpine:latest
# Install curl
RUN apk update && apk add curl

# Set the Current Working Directory inside the container
WORKDIR /root/

RUN mkdir ./keys

# 复制对应文件
COPY web ./web

COPY upload.sh  ./

COPY novel-x86  ./

COPY novel-arm  ./

# Expose port 8080 to the outside world
EXPOSE 3388

# Command to run the executable
CMD ["./novel-arm"]