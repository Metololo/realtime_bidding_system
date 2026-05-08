FROM golang:alpine

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o seller-service ./cmd/seller-service

CMD ["./seller-service"]
