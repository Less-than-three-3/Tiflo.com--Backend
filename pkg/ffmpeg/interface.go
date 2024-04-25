package ffmpeg

import (
	"tiflo/model"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type MediaService interface {
	SplitAudio(audioPartToSplit model.AudioPart, splitPointStr string, duration time.Duration) ([]model.AudioPart, error)
	ExtractFrame(videoPath string, timestamp string) (uuid.UUID, error)

	ConvertTimeFromString(timeString string) int64
	GetAudioDuration(audioPath string) (time.Duration, int64, error)
}

type MediaServiceImpl struct {
	pathForMedia string
	logger       *logrus.Entry
}

func NewMediaService(pathForMedia string, logger *logrus.Logger) MediaService {
	return &MediaServiceImpl{pathForMedia: pathForMedia, logger: logger.WithField("component", "media-service")}
}
