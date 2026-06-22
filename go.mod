module test

go 1.26.3

replace study/contracts => ./contracts

require (
	google.golang.org/grpc v1.81.1
	study/contracts v0.0.0-00010101000000-000000000000
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
