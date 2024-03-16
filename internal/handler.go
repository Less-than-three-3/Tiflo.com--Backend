package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Handler struct {
	logger *logrus.Entry
}

func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger: logger.WithField("component", "handler"),
	}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	return r
}
