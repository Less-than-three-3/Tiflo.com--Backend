package internal

import (
	"net/http"
	"tiflo/model"
	"tiflo/pkg/grpc/client"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger       *logrus.Entry
	pythonClient client.AI
}

func NewHandler(logger *logrus.Logger, pythonClient client.AI) *Handler {
	return &Handler{
		logger:       logger.WithField("component", "handler"),
		pythonClient: pythonClient,
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (h *Handler) InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/voice-to-text", h.GetVoicedText)
	}

	return r
}

func (h *Handler) GetVoicedText(context *gin.Context) {
	var textComment model.TextToVoice
	if err := context.BindJSON(&textComment); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	audioBytes, err := h.pythonClient.VoiceTheText(context.Request.Context(), textComment)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	context.Writer.Header().Set("Content-Type", "audio/wav")
	context.Writer.Write(audioBytes)
}
