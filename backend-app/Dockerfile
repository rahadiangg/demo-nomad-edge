FROM golang:1.22.1-alpine as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o backend-app .

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/backend-app ./
CMD [ "/app/backend-app" ]