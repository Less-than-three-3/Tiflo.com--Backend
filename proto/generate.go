package proto_gen

//go:generate protoc  --go_out=..  --go-grpc_out=.. --proto_path=. voice2text.proto
//go:generate protoc  --go_out=..  --go-grpc_out=.. --proto_path=. image2text.proto
