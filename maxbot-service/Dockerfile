FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH=$PATH:/root/go/bin

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/maxbot.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/maxbot-service ./cmd/maxbot

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/maxbot-service .

EXPOSE 9095

CMD ["./maxbot-service"]
