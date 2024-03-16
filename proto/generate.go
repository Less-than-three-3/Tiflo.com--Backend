package proto_gen

//go:generate protoc  --go_out=..  --go-grpc_out=.. --proto_path=. model.proto
