# 构建阶段
FROM golang:1.24.0-alpine AS builder

# 设置Go代理
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o WgInspector

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/WgInspector .
# 复制配置文件
COPY app/init_config.yaml ./app/
COPY app/config ./app/config

# 设置环境变量
ENV TZ=Asia/Shanghai

# 暴露端口（如果需要）
EXPOSE 9999

# 设置入口点
ENTRYPOINT ["./WgInspector"]
