package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-audio/wav"
	"github.com/google/uuid"
	"net/http"
	"os"
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

	path, err := h.pythonClient.VoiceTheText(context.Request.Context(), text)
	if err != nil {
		h.logger.Error(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	//path := "test.wav"
	// get duration

	file, err := os.Open(PathForMedia + path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	duration, err := decoder.Duration()

	// split with duration |   |       |    |
	//                         |comment|

	// ffmpeg -i demo.mp3 -vn -acodec copy -ss 00:00:00 -t 00:01:50 demo-cut.mp3
	// split
	// get first part
	firstPartName := uuid.New()
	fmt.Println("ffmpeg", "-i", PathForMedia+project.AudioPath, "-vn", "-acodec", "copy",
		"-ss", "00:00:00", "-t", fmt.Sprintf("%02d:%02d:%02d", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60),
		firstPartName.String()+".mp3")

	_, err = exec.Command("ffmpeg", "-i", PathForMedia+project.AudioPath, "-vn", "-acodec", "copy",
		"-ss", "00:00:00", "-t", fmt.Sprintf("%02d:%02d:%02d", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60),
		firstPartName.String()+".mp3").Output()
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// get second part
	secondPartName := uuid.New()
	formattedDuration := fmt.Sprintf("%02d:%02d:%02d", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60+int(h.convertTime(comment.Start)))

	fmt.Println("ffmpeg", "-i", PathForMedia+project.AudioPath, "-vn", "-acodec", "copy",
		"-ss", formattedDuration, secondPartName.String()+".mp3")

	_, err = exec.Command("ffmpeg", "-i", PathForMedia+project.AudioPath, "-vn", "-acodec", "copy",
		"-ss", formattedDuration, secondPartName.String()+".mp3").Output()
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	fileAll, err := os.Open(PathForMedia + path)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer file.Close()

	decoder = wav.NewDecoder(fileAll)
	durationAll, err := decoder.Duration()
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	secondStart := h.convertTime(comment.Start) + int64(duration.Seconds())
	project.AudioParts = []model.AudioPart{
		{
			PartId:    uuid.New(),
			ProjectId: projectId,
			Start:     0,
			Duration:  h.convertTime(comment.Start),
			Text:      "",
			Path:      PathForMedia + firstPartName.String() + ".mp3",
		}, {
			PartId:    uuid.New(),
			ProjectId: projectId,
			Start:     h.convertTime(comment.Start) + int64(duration.Seconds()),
			Duration:  int64(durationAll.Seconds()) - secondStart,
			Text:      "",
			Path:      PathForMedia + secondPartName.String() + ".mp3",
		},
		{
			PartId:    uuid.New(),
			ProjectId: projectId,
			Start:     h.convertTime(comment.Start),
			Duration:  int64(duration.Seconds()),
			Text:      text,
			Path:      PathForMedia + path,
		},
	}

	err = h.repo.SaveProjectAudio(context.Request.Context(), project)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, project)
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
