package main

import (
	"time"

	"tiflo/internal/handler"

	"github.com/sirupsen/logrus"
)

// @title Tiflo_Backend
// @version 1.0
// @description App for working with audio descriptions(tiflocomments)

// @host http://158.160.2.243:8080
// @schemes http https
// @BasePath /
func main() {
	logger := logrus.New()
	formatter := &logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	}
	logger.SetFormatter(formatter)

	handler := handler.NewHandler(logger)
	r := handler.InitRouter()
	r.Run("0.0.0.0:8080")
}
