package ffmpeg

import (
	"os/exec"

	"github.com/google/uuid"
)

// ExtractFrame gets frame from mp4 file as png and returns its name
func (s *MediaServiceImpl) ExtractFrame(videoPath string, timestamp string) (string, error) {
	var frameName = uuid.New()
	_, err := exec.Command("ffmpeg", "-i", videoPath, "-ss", timestamp, "-frames:v",
		"1", s.pathForMedia+frameName.String()+".png").Output()
	if err != nil {
		s.logger.Error("error while extracting frame: ", err)
		return "", err
	}

	return frameName.String() + ".png", nil
}
