package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"tiflo/model"
)

// CreateComment godoc
// @Summary      Create comment on video
// @Description  Create comment on video using split point
// @Tags         Comment
// @Accept       json
// @Produce      json
// @Param        projectId  path  string  true  "Project Id"
// @Param        comment  body  model.Comment  true  "Split point"
// @Success      200  {object}  model.Project
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId}/video/comment [post]
func (h *Handler) CreateComment(context *gin.Context) {
	projectIdStr := context.Param("projectId")
	projectId, err := uuid.Parse(projectIdStr)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	var comment model.Comment
	if err = context.BindJSON(&comment); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "неверный формат данных")
		return
	}

	project, err := h.repo.GetProject(context.Request.Context(), model.Project{ProjectId: projectId, UserId: userId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	frameName, err := h.mediaService.ExtractFrame(PathForMedia+project.VideoPath, comment.VideoTime)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	text, err := h.pythonClient.ImageToText(context.Request.Context(), "/data/"+frameName)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	path, err := h.pythonClient.VoiceTheText(context.Request.Context(), text)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// get duration

	duration, durationInt, err := h.mediaService.GetAudioDurationWav(path)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	splitPoint := h.mediaService.ConvertTimeFromString(comment.SplitPoint)

	audioPartToSplit, err := h.repo.GetAudioPartBySplitPoint(context.Request.Context(), splitPoint, projectId)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	splittedParts, err := h.mediaService.SplitAudio(audioPartToSplit, comment.SplitPoint, duration)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	_, err = h.repo.DeleteAudioPart(context.Request.Context(), model.AudioPart{PartId: audioPartToSplit.PartId, ProjectId: projectId})
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	splittedParts = append(splittedParts, model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: projectId,
		Start:     splitPoint,
		Duration:  durationInt,
		Text:      text,
		Path:      path,
	})

	audioPartsAfterSplitPoint, err := h.repo.GetAudioPartsAfterSplitPoint(context.Request.Context(), splitPoint, projectId)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	for i, _ := range audioPartsAfterSplitPoint {
		audioPartsAfterSplitPoint[i].Start += durationInt
	}

	audioPartsAfterSplitPoint = append(audioPartsAfterSplitPoint, splittedParts...)

	for _, part := range audioPartsAfterSplitPoint {
		err = h.repo.UpdateAudioPart(context.Request.Context(), part)
		if err != nil {
			h.logger.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	updatedProject, err := h.repo.GetProject(context.Request.Context(), model.Project{ProjectId: projectId, UserId: userId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, updatedProject)
}
