package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"tiflo/model"
)

func (h *Handler) DeleteAudioPart(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	audioPartIdStr := context.Param("audioPartId")
	audioPartId, err := uuid.Parse(audioPartIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	err = h.repo.DeleteAudioPart(context.Request.Context(), model.AudioPart{PartId: audioPartId, ProjectId: projectId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}
