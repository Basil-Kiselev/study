IMAGE_NAME = proto-builder
PROJECT_ROOT = $(shell pwd)/..
WORKDIR = /app

.PHONY: build-image
build-image:
	docker build -t $(IMAGE_NAME) .

.PHONY: generate
generate: build-image
	@echo "Generating proto files..."
	docker run --rm -v $(PROJECT_ROOT):$(WORKDIR) -w $(WORKDIR) $(IMAGE_NAME) \
		protoc \
		-I contracts \
		-I /usr/local/include/googleapis \
		--go_out=contracts --go_opt=paths=source_relative \
		--go-grpc_out=contracts --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=contracts --grpc-gateway_opt=paths=source_relative \
		contracts/pagination/pagination.proto \
		contracts/account/account_model.proto \
		contracts/account/account_service.proto \
		contracts/auth/auth_service.proto \
		contracts/gateway/gateway_service.proto \
		contracts/transaction/transaction_model.proto \
		contracts/transaction/transaction_service.proto
	@echo "Done!"

.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	find . -name "*.pb.go" -delete
	find . -name "*.pb.gw.go" -delete
