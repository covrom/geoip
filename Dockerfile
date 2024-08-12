FROM golang:alpine as builder

RUN mkdir -p /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o geoip .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/geoip .

CMD ["/app/geoip"]
