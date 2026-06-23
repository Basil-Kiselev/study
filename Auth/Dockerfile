FROM golang:1.26-alpine as build
WORKDIR /app

COPY Auth/go.mod Auth/go.sum ./Auth/
COPY contracts/go.mod contracts/go.sum ./contracts/

WORKDIR /app/Auth
RUN go mod download

WORKDIR /app
COPY Auth/ ./Auth/
COPY contracts/ ./contracts/

WORKDIR /app/Auth
RUN go build -o app ./cmd

FROM alpine:3.10
WORKDIR /app

COPY --from=build /app/Auth/app .
COPY --from=build /app/Auth/internal/migrations ./internal/migrations

CMD [ "./app" ]
