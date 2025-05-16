package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"tiflo/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	userId, err := model.GetUserId(context)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	project, err := h.repo.GetProject(context.Request.Context(), model.Project{ProjectId: projectId, UserId: userId})
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	sort.SliceStable(project.AudioParts, func(i, j int) bool {
		return project.AudioParts[i].Start < project.AudioParts[j].Start
	})

	partsToConcat := make([]model.AudioPart, 0, 2)
	for i, v := range project.AudioParts {
		if v.PartId == audioPartId && i != 0 {
			partsToConcat = append(partsToConcat, project.AudioParts[i-1], project.AudioParts[i+1])
			h.logger.Info("partsToConcat:", partsToConcat)

			path, err := h.mediaService.ConcatAudio(partsToConcat)
			if err != nil {
				context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}

			concatedPart := model.AudioPart{
				PartId:    project.AudioParts[i-1].PartId,
				ProjectId: projectId,
				Start:     project.AudioParts[i-1].Start,
				Duration:  project.AudioParts[i-1].Duration + project.AudioParts[i+1].Duration,
				Text:      "",
				Path:      path,
			}

			_, err = h.repo.DeleteAudioPart(context.Request.Context(), project.AudioParts[i+1])
			if err != nil {
				context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}

			_, err = h.repo.DeleteAudioPart(context.Request.Context(), v)
			if err != nil {
				context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}

			audioPartsAfterSplitPoint, err := h.repo.GetAudioPartsAfterSplitPoint(context.Request.Context(), v.Start, projectId)
			if err != nil {
				h.logger.Error(err)
				context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}

			for j, _ := range audioPartsAfterSplitPoint {
				audioPartsAfterSplitPoint[j].Start -= v.Duration
			}

			audioPartsAfterSplitPoint = append(audioPartsAfterSplitPoint, concatedPart)
			for _, part := range audioPartsAfterSplitPoint {
				err = h.repo.UpdateAudioPart(context.Request.Context(), part)
				if err != nil {
					h.logger.Error(err)
					context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
					return
				}
			}

			break
		}
	}

	context.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}

// ChangeCommentText godoc
// @Summary      Change text comment
// @Description  Change text comment for chosen audio part
// @Tags         Audio part
// @Accept       json
// @Produce      json
// @Param        projectId  path  string  true  "Project Id"
// @Param        audioPartId  path  string  true  "Audio part Id"
// @Param        comment  body  model.Comment  true  "New text for comment"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  error
// @Failure      401  {object}  error
// @Failure      500  {object}  error
// @Router       /api/projects/{projectId}/audio-part/{audioPartId} [put]
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
	path, err := h.ttsClient.TextToSpeech(context.Request.Context(), comment.Text)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// get duration of new text
	_, durationInt, err := h.mediaService.GetAudioDurationMp3(path)
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

func (h *Handler) UploadAudio(context *gin.Context) {
	h.logger.Info("UploadAudio")

	// Получаем файл из формы
	file, err := context.FormFile("voice")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get audio file: " + err.Error(),
		})
		return
	}

	// Создаем папку для сохранения, если не существует
	uploadDir := "/media"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create upload directory: " + err.Error(),
		})
		return
	}

	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String()
	fullFilename := filename + ext
	filePath := filepath.Join(uploadDir, fullFilename)

	// Сохраняем файл
	src, err := file.Open()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open uploaded file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	out, err := os.Create(filePath)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create file on server: " + err.Error(),
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file as webm: " + err.Error(),
		})
		return
	}

	h.logger.Info("WEBM file successfully saved")

	wavFilename := filename + ".wav"
	wavFilepath := uploadDir + "/" + wavFilename
	err = h.mediaService.ConvertWebmToWav(filePath, wavFilepath)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to save file as wav: " + err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message":  "Audio file uploaded successfully",
		"filepath": wavFilename,
	})
}
