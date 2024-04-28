package ffmpeg

import "os/exec"

func (s *MediaServiceImpl) GetAudioFromVideo(filename string, extension string) error {
	_, err := exec.Command("ffmpeg", "-i", s.pathForMedia+filename+extension, s.pathForMedia+filename+".wav").Output()
	if err != nil {
		s.logger.Error(err)
		return err
	}

	return nil
}
