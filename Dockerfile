# --- Build stage ---
    FROM golang:1.23 AS builder
    # ถ้าอยากกันพลาดเรื่อง toolchain อีก ใส่บรรทัดนี้เพิ่มได้:
    ENV GOTOOLCHAIN=auto
    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # สร้าง Swagger docs (ให้ตรงกับ lib)
    RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3
    RUN /go/bin/swag init -g cmd/main.go -o docs
    
    RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd
    
    # --- Run stage ---
    FROM gcr.io/distroless/base-debian12
    ENV APP_PORT=8080
    WORKDIR /app
    COPY --from=builder /app/server /app/server
    COPY --from=builder /app/docs /app/docs
    EXPOSE 8080
    USER nonroot:nonroot
    ENTRYPOINT ["/app/server"]
    