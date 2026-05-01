FROM golang:1.24

# Устанавливаем protoc
ARG PROTOC_VERSION=27.1
RUN apt-get update && apt-get install -y unzip curl git && \
    ARCH=$(uname -m) && \
    if [ "$ARCH" = "aarch64" ]; then PROTOC_ARCH="linux-aarch_64"; else PROTOC_ARCH="linux-x86_64"; fi && \
    curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${PROTOC_ARCH}.zip -o /tmp/protoc.zip && \
    unzip /tmp/protoc.zip -d /usr/local && \
    rm -rf /tmp/protoc.zip /var/lib/apt/lists/*

# Качаем googleapis
RUN git clone --depth=1 https://github.com/googleapis/googleapis /usr/local/include/googleapis

# Устанавливаем плагины по одному (так надежнее и понятнее при ошибках)
ENV GOPROXY=https://proxy.golang.org,direct

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.22.0
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.22.0

ENV PATH="$PATH:/go/bin"
WORKDIR /app
