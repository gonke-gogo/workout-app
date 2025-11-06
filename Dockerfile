FROM golang:1.21-alpine AS builder

RUN apk add --no-cache protobuf-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

COPY . .

RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/workout.proto

# cmd/server/main.goのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 50051

CMD ["./main"]
