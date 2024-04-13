package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"tiflo/model"
)

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

	var imageName = uuid.New()
	// extract frame
	_, err = exec.Command("ffmpeg", "-i", PathForMedia+project.VideoPath, "-ss", comment.Start, "-frames:v",
		"1", PathForMedia+imageName.String()+".png").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}

	// send to python voice path=PathForMedia+imageName.String()+".png"

	text := "группа мужчин, стоящих рядом с черной машиной. Они одеты в синюю форму, и автомобиль кажется BMW. Мужчины расположены перед машиной, а сцена происходит на грунтовой дороге."

	//path, err := h.pythonClient.VoiceTheText(context.Request.Context(), text)
	//if err != nil {
	//	h.logger.Error(err)
	//	context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	//	return
	//}

	path := "test.wav"

	err = h.repo.AddAudioPart(context.Request.Context(), model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: projectId,
		Start:     h.convertTime(comment.Start),
		Duration:  0,
		Text:      text,
		Path:      PathForMedia + path,
	})

	context.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (h *Handler) convertTime(timeString string) int64 {
	parts := strings.Split(timeString, ":")
	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		h.logger.Error(err)
	}
	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		h.logger.Error(err)
	}

	return int64(minutes*60 + seconds)
}
