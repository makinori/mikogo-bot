FROM golang:1.24.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o mikogo-mumble

FROM alpine:latest

# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN apk add --no-cache chromium
ENV CHROMIUM_PATH=/usr/bin/chromium

WORKDIR /app

COPY --from=build /app/mikogo-mumble /app/mikogo-mumble
COPY assets/ /app/assets/

ENTRYPOINT ["/app/mikogo-mumble"]