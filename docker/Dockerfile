# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
  && adduser -D -H -s /sbin/nologin helios
WORKDIR /app
COPY --from=build /out/server ./server
USER helios
EXPOSE 8080
ENTRYPOINT ["./server"]
