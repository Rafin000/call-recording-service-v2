# Build Stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final Stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser

COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .

RUN chown -R appuser:appgroup .
USER appuser

EXPOSE 8081
CMD ["./main"]
