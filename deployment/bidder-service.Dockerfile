FROM golang:alpine

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bidder-service ./cmd/bidder-service

CMD ["./bidder-service"]
