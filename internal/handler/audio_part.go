package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"tiflo/model"
)

// DeleteAudioPart godoc
// @Summary      Delete audio part
// @Description  Delete audio part
// @Tags         Audio part
// @Produce      json
// @Param        projectId    path  string  true  "Project Id"
// @Param        audioPartId  path  string  true  "Audio part Id"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId}/audio-part/{audioPartId} [delete]
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

	part, err := h.repo.DeleteAudioPart(context.Request.Context(), model.AudioPart{PartId: audioPartId, ProjectId: projectId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	audioPartsAfterSplitPoint, err := h.repo.GetAudioPartsAfterSplitPoint(context.Request.Context(), part.Start, projectId)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for i, _ := range audioPartsAfterSplitPoint {
		audioPartsAfterSplitPoint[i].Start -= part.Duration
	}

	for _, part := range audioPartsAfterSplitPoint {
		err = h.repo.UpdateAudioPart(context.Request.Context(), part)
		if err != nil {
			h.logger.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	context.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}

func (h *Handler) ChangeCommentText(context *gin.Context) {
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

	var comment model.Comment
	if err = context.BindJSON(&comment); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "неверный формат данных")
		return
	}

	// delete all audio part and delete its duration from parts which go after it
	oldPart, err := h.repo.DeleteAudioPart(context.Request.Context(), model.AudioPart{PartId: audioPartId, ProjectId: projectId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	audioPartsAfterSplitPoint, err := h.repo.GetAudioPartsAfterSplitPoint(context.Request.Context(), oldPart.Start, projectId)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for i, _ := range audioPartsAfterSplitPoint {
		audioPartsAfterSplitPoint[i].Start -= oldPart.Duration
	}

	// voice new text
	path, err := h.pythonClient.VoiceTheText(context.Request.Context(), comment.Text)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// get duration of new text
	_, durationInt, err := h.mediaService.GetAudioDurationWav(path)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// add new part duration to parts which go after
	for i, _ := range audioPartsAfterSplitPoint {
		audioPartsAfterSplitPoint[i].Start += durationInt
	}

	audioPartsAfterSplitPoint = append(audioPartsAfterSplitPoint, model.AudioPart{
		PartId:    oldPart.PartId,
		ProjectId: oldPart.ProjectId,
		Start:     oldPart.Start,
		Duration:  durationInt,
		Text:      comment.Text,
		Path:      path,
	})

	for _, part := range audioPartsAfterSplitPoint {
		err = h.repo.UpdateAudioPart(context.Request.Context(), part)
		if err != nil {
			h.logger.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	context.JSON(http.StatusOK, gin.H{"message": "successfully changed"})
}
