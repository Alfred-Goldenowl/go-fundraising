FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["/app/main"]
