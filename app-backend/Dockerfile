FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go


FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]