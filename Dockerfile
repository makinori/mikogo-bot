FROM golang:1.24.1 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o mikogo-irc-bot

FROM scratch

WORKDIR /app

COPY --from=build /app/mikogo-irc-bot /app/mikogo-irc-bot

ENTRYPOINT ["/app/mikogo-irc-bot"]