FROM gocv/opencv:4.9.0 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN make build

# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
FROM gocv/opencv:4.9.0
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/bin/monitoring-system.out /usr/local/bin/camera-monitor

WORKDIR /app

COPY --from=builder /app/src/web/static /app/src/web/static

EXPOSE 4000

CMD ["camera-monitor"]
