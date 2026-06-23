FROM golang:1.26-alpine as build
WORKDIR /app

COPY Account/go.mod Account/go.sum ./Account/
COPY contracts/go.mod contracts/go.sum ./contracts/

WORKDIR /app/Account
RUN go mod download

WORKDIR /app
COPY Account/ ./Account/
COPY contracts/ ./contracts/

WORKDIR /app/Account
RUN go build -o app ./cmd

FROM alpine:3.10
WORKDIR /app

COPY --from=build /app/Account/app .
COPY --from=build /app/Account/internal/migrations ./internal/migrations

CMD [ "./app" ]
