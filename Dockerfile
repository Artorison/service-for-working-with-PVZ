FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /myWorkDir

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/pvz cmd/main.go

FROM alpine:3.21 AS runner
WORKDIR /myWorkDir

COPY --from=builder /myWorkDir/bin/pvz .
COPY config/config.yml config/config.yml

CMD ["./pvz"]