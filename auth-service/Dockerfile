# build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# install git for go modules if private modules are used
RUN apk add --no-cache git protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH=$PATH:/root/go/bin

# Copy the entire project (needed for replace directives)
COPY . .

# Change to auth-service directory
WORKDIR /app/auth-service

# Download dependencies (replace directives will work now)
RUN go mod download

# Skip proto generation for now - using existing generated files

# Skip swagger generation for now

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/auth

# final
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /auth-service .
EXPOSE 8080
ENV HTTP_ADDR=:8080
CMD ["./auth-service"]