package internal

import (
	"net/http"
	"tiflo/model"
	"tiflo/pkg/grpc/client"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const PathForMedia = "/media/"

type Handler struct {
	logger       *logrus.Entry
	pythonClient client.AI
	repo         Repository
}

func NewHandler(logger *logrus.Logger, repo Repository) *Handler {
	return &Handler{
		logger: logger.WithField("component", "handler"),
		//pythonClient: pythonClient,
		repo: repo,
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
		apiGroup.POST("/voice-the-text", h.GetVoicedText)
		apiGroup.POST("/save-image", h.SaveImage)
		apiGroup.GET("/get-image-project", h.GetImageProject)
	}

	return r
}

func (h *Handler) SaveImage(context *gin.Context) {
	form, _ := context.MultipartForm()
	files := form.File["file"]

	for _, file := range files {
		if err := context.SaveUploadedFile(file, PathForMedia+file.Filename); err != nil {
			h.logger.Printf("Failed to save file: %s", err)
			context.String(500, "Failed to save file")
			return
		}

		if err := h.repo.SaveImageProject(context.Request.Context(), model.ImageProject{Image: file.Filename}); err != nil {
			context.String(500, "Failed to save file")
			return
		}
	}

	context.String(200, "File uploaded successfully")
}

func (h *Handler) GetImageProject(context *gin.Context) {
	project, err := h.repo.GetImageProject(context.Request.Context(), model.ImageProject{})
	if err != nil {
		context.String(500, err.Error())
		return
	}
	context.JSON(http.StatusOK, project)
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
	_, err = context.Writer.Write(audioBytes)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
}
