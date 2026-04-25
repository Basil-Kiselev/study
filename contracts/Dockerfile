FROM golang:1.24

# Устанавливаем protoc
ARG PROTOC_VERSION=27.1
RUN apt-get update && apt-get install -y unzip curl git && \
    rm -rf /var/lib/apt/lists/* && \
    curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip -o /tmp/protoc.zip && \
    unzip /tmp/protoc.zip -d /usr/local && \
    rm /tmp/protoc.zip

# Качаем googleapis для поддержки google.api.http (annotations.proto)
RUN git clone --depth=1 https://github.com/googleapis/googleapis /usr/local/include/googleapis

# Устанавливаем плагины gRPC и Gateway
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Добавляем пути к бинарникам Go
ENV PATH="$PATH:/go/bin"

WORKDIR /app
