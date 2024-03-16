package main

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"tiflo/internal"
	pythonClient "tiflo/pkg/grpc/client"
	pb "tiflo/pkg/grpc/generated"

	"time"
)

func main() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logger.SetFormatter(formatter)

	addr := "172.21.133.141:8080"
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	logger.Info("connected to python")

	client := pb.NewAIServiceClient(conn)
	handler := internal.NewHandler(logger, pythonClient.NewPythonClient(logger, client))

	r := handler.InitRouter()
	r.Run("0.0.0.0:8080")
}
