# 使用官方Go镜像作为构建环境
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod文件
COPY go.mod ./

# 复制源代码（移动到这里，以便 tidy 可以扫描代码）
COPY . .

RUN mkdir -p output

# 下载依赖，现在 tidy 可以正确工作
RUN go mod tidy && go mod download

# 构建应用
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
# 构建armv7版本
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o main ./cmd/main.go

# 使用轻量级的alpine镜像作为运行环境
FROM scratch

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/output /output

# 复制时区信息
COPY ./usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Shanghai

# 运行应用
CMD ["./main"]

# 本地构建
# docker build -t exporter-to-obsidian:latest .  

# docker buildx build --platform linux/arm/v7 -t exporter_to_obsidian:latest .

# 推送到Docker Hub
# docker buildx build --platform linux/arm64,linux/amd64 -t username/exporter-to-obsidian:latest --push .

# 推送armv7版本
# docker buildx build --platform linux/arm/v7 -t username/exporter-to-obsidian:armv7 --push .