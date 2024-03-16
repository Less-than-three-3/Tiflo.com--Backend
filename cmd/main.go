package main

import (
	"github.com/sirupsen/logrus"
	"tiflo/internal"
	"time"
)

func main() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logger.SetFormatter(formatter)

	handler := internal.NewHandler(logger)
	r := handler.InitRouter()
	r.Run("0.0.0.0:8080")
}
