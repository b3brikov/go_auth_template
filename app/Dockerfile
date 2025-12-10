FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN ls -la
RUN go mod download
COPY . .
RUN ls -la /app/cmd/auth/
RUN ls -la /app/config/
RUN go build cmd/auth/main.go
CMD ["./main", "--config=./config/local.yaml"]