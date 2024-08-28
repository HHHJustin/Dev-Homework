# 使用 Go 的官方映像版本
FROM golang:1.21.6

# 設定工作目錄到 /app
WORKDIR /app

# 複製所有文件到容器中（包含 go.mod 和 go.sum）
COPY . .

# 下載所有依賴項
RUN go mod tidy

# 設定環境變數
ENV PORT=9999

# 執行應用程式
CMD ["go", "run", "main.go"]
