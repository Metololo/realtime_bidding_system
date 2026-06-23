FROM golang:alpine

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o auction-engine ./cmd/auction-engine

EXPOSE 8080 9001

CMD ["./auction-engine"]