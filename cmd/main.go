package main

import (
	"tiflo/internal"
	"time"

	"github.com/sirupsen/logrus"
)

// @title Tiflo_Backend
// @version 1.0
// @description App for working with audio descriptions(tiflocomments)

// @host localhost:8080
// @schemes http
// @BasePath /
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
