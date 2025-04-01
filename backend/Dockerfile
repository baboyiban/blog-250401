FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p db

# gcc 설치
RUN apk update && apk add --no-cache gcc musl-dev

# CGO 활성화
RUN CGO_ENABLED=1 go build -o main ./src/cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/db/* ./db/

EXPOSE 8080

CMD ["./main"]
