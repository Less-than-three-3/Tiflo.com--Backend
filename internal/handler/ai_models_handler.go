package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"tiflo/model"
)

// VoiceText godoc
// @Summary      Voice the given text
// @Description  Voice the given text
// @Tags         Project
// @Accept       json
// @Produce      json
// @Param        user  body  model.VoiceText  true  "text which you want to be voiced"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId}/voice [post]
func (h *Handler) VoiceText(context *gin.Context) {
	var textComment model.VoiceText

	if err := context.BindJSON(&textComment); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	h.logger.Info("VoiceText Handler", textComment.Text)

	path, err := h.pythonClient.VoiceTheText(context.Request.Context(), textComment.Text)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	context.JSON(http.StatusOK, path)
}

// ImageToText godoc
// @Summary      Create tiflo comment
// @Description  Create tiflo comment for given image
// @Tags         Comment
// @Accept       json
// @Produce      json
// @Param        user  body  model.Image  true  "name of image(uuid)"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId}/image/comment [post]
func (h *Handler) ImageToText(context *gin.Context) {
	var imagePath model.Image
	h.logger.Info(imagePath)

	if err := context.BindJSON(&imagePath); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	text, err := h.pythonClient.ImageToText(context.Request.Context(), "/home/vavasto/frontend/public/media/"+imagePath.Name)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	h.logger.Info(text)

	err = h.repo.SaveProjectAudio(context.Request.Context(), model.Project{
		ProjectId: projectId,
		UserId:    uuid.UUID{},
		AudioParts: []model.AudioPart{{
			PartId:    uuid.New(),
			ProjectId: projectId,
			Start:     0,
			Duration:  0,
			Text:      text,
			Path:      "",
		}},
	})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	context.JSON(http.StatusOK, text)
}
