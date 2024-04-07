package handler

import (
	"github.com/gin-gonic/gin"
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
	h.logger.Info(textComment)

	if err := context.BindJSON(&textComment); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	path, err := h.pythonClient.VoiceTheText(context.Request.Context(), textComment.Text)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	context.JSON(http.StatusOK, path)
}
