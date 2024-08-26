# 使用官方的Go镜像作为构建环境
FROM golang:1.20 AS builder

# 设置工作目录
WORKDIR /app


# 将Go的模块文件和源代码复制到工作目录，主要是为了利用Docker的缓存机制
COPY go.mod go.sum ./
RUN go mod tidy

# 复制所有源代码到工作目录
COPY . .

# 切换到internal/starter目录
WORKDIR /app/internal/starter

# 构建Go应用并指定输出文件名为offercat-go
RUN go build -o /app/bin/offercat-go

# 使用一个轻量级的Alpine镜像作为运行环境
FROM alpine:3.18


# 设置时区（可选）
RUN apk add --no-cache tzdata

# 从构建环境中复制应用程序
COPY --from=builder /app/bin/offercat-go /usr/local/bin/offercat-go

# 暴露服务端口
EXPOSE 12345

# 启动应用
CMD ["offercat-go"]
