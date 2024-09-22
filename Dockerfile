FROM gocv/opencv:4.9.0 as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go install github.com/joho/godotenv/cmd/godotenv@latest

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags no_gpu -o camera-monitor ./cmd/main.go

# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
FROM gocv/opencv:4.9.0
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/camera-monitor /usr/local/bin/camera-monitor

COPY .env /app/.env

WORKDIR /app

EXPOSE 4000

CMD ["camera-monitor"]