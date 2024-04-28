package ffmpeg

import (
	"github.com/google/uuid"
	"tiflo/model"
	"time"

	"github.com/sirupsen/logrus"
)

type MediaService interface {
	SplitAudio(audioPartToSplit model.AudioPart, splitPointStr string, duration time.Duration) ([]model.AudioPart, error)
	ConcatAudio(audioParts []model.AudioPart) (string, error)

	ConvertTimeFromString(timeString string) int64

	GetAudioDurationWav(audioPath string) (time.Duration, int64, error)
	GetAudioDurationMp3(audioPath string) (time.Duration, int64, error)

	GetAudioFromVideo(filename string, extension string) error
	ExtractFrame(videoPath string, timestamp string) (uuid.UUID, error)
}

type MediaServiceImpl struct {
	pathForMedia string
	logger       *logrus.Entry
}

func NewMediaService(pathForMedia string, logger *logrus.Logger) MediaService {
	return &MediaServiceImpl{pathForMedia: pathForMedia, logger: logger.WithField("component", "media-service")}
}
