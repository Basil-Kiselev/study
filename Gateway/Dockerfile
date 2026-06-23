FROM golang:1.26-alpine as build
WORKDIR /app

COPY Gateway/go.mod Gateway/go.sum ./Gateway/
COPY contracts/go.mod contracts/go.sum ./contracts/

WORKDIR /app/Gateway
RUN go mod download

WORKDIR /app
COPY Gateway/ ./Gateway/
COPY contracts/ ./contracts/

WORKDIR /app/Gateway
RUN go build -o app ./cmd

FROM alpine:3.10
WORKDIR /app

COPY --from=build /app/Gateway/app .

CMD [ "./app" ]
