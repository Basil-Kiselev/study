FROM golang:1.26-alpine as build
WORKDIR /app

COPY Transaction/go.mod Transaction/go.sum ./Transaction/
COPY contracts/go.mod contracts/go.sum ./contracts/

WORKDIR /app/Transaction
RUN go mod download

WORKDIR /app
COPY Transaction/ ./Transaction/
COPY contracts/ ./contracts/

WORKDIR /app/Transaction
RUN go build -o app ./cmd

FROM alpine:3.10
WORKDIR /app

COPY --from=build /app/Transaction/app .
COPY --from=build /app/Transaction/internal/migrations ./internal/migrations

CMD [ "./app" ]
