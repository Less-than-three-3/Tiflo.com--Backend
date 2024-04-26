package ffmpeg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"tiflo/model"

	"github.com/go-audio/wav"
	"github.com/google/uuid"
	"github.com/tcolgate/mp3"
)

// SplitAudio splits audioPart in two parts and recount their start and duration according to duration of voiced tiflo comment
// splitPoint comes in format hh:mm:ss.ms (ms takes 3 signs)
// before
//
//	start       splitPoint             end1
//	     |           |                  |
//
// after
//
//	start       splitPoint  duration1+splitPoint         duration1+end1
//	     |           |             |                  |
//	                 |  duration1  |
func (s *MediaServiceImpl) SplitAudio(audioPartToSplit model.AudioPart, splitPointStr string, duration time.Duration) ([]model.AudioPart, error) {
	var result = make([]model.AudioPart, 0, 2)

	firstPartName := uuid.New()

	start := audioPartToSplit.Start
	startStr := s.convertTimeToString(start)

	splitPoint := s.ConvertTimeFromString(splitPointStr)
	s.logger.Info("splitPoint: ", splitPoint)

	s.logger.Info("ffmpeg", "-i", audioPartToSplit.Path, "-vn", "-acodec", "copy",
		"-ss", startStr, "-t", fmt.Sprintf("%02d:%02d:%02d.%03d", int(duration.Hours()), int(duration.Minutes())%60,
			int(duration.Seconds())%60, int(duration.Milliseconds())%60),
		s.pathForMedia+firstPartName.String()+".mp3")

	_, err := exec.Command("ffmpeg", "-i", audioPartToSplit.Path, "-vn", "-acodec", "copy",
		"-ss", startStr, "-t", fmt.Sprintf("%02d:%02d:%02d.%03d", int(duration.Hours()), int(duration.Minutes())%60,
			int(duration.Seconds())%60, int(duration.Milliseconds())%60),
		s.pathForMedia+firstPartName.String()+".mp3").Output()
	if err != nil {
		return nil, err
	}
	result = append(result, model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: audioPartToSplit.ProjectId,
		Start:     audioPartToSplit.Start,
		Duration:  splitPoint - audioPartToSplit.Start,
		Text:      "",
		Path:      firstPartName.String(),
	})

	secondPartName := uuid.New()
	s.logger.Info("ffmpeg", "-i", audioPartToSplit.Path, "-vn", "-acodec", "copy",
		"-ss", s.convertTimeToString(splitPoint+s.convertDurationToInt64(duration)), "-t", s.convertTimeToString(audioPartToSplit.Duration-splitPoint),
		s.pathForMedia+secondPartName.String()+".mp3")

	_, err = exec.Command("ffmpeg", "-i", audioPartToSplit.Path, "-vn", "-acodec", "copy",
		"-ss", s.convertTimeToString(splitPoint+s.convertDurationToInt64(duration)), "-t", s.convertTimeToString(audioPartToSplit.Duration-splitPoint),
		s.pathForMedia+secondPartName.String()+".mp3").Output()
	if err != nil {
		return nil, err
	}

	result = append(result, model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: audioPartToSplit.ProjectId,
		Start:     splitPoint + s.convertDurationToInt64(duration),
		Duration:  audioPartToSplit.Duration - splitPoint,
		Text:      "",
		Path:      secondPartName.String(),
	})

	return result, nil
}

func (s *MediaServiceImpl) GetDuration(filename string) (time.Duration, error) {
	file, err := os.Open(s.pathForMedia + filename)
	if err != nil {
		s.logger.Error("error while getting duration : ", err)
		return time.Duration(0), err
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	return decoder.Duration()
}

// timeString comes in format hh:mm:ss.ms
func (s *MediaServiceImpl) ConvertTimeFromString(timeString string) int64 {
	parts := strings.Split(timeString, ":")
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		s.logger.Error(err)
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		s.logger.Error(err)
	}

	secondsWithMillisecondsArr := strings.Split(parts[2], ".")
	seconds, err := strconv.Atoi(secondsWithMillisecondsArr[0])
	if err != nil {
		s.logger.Error(err)
	}

	milliseconds, err := strconv.Atoi(secondsWithMillisecondsArr[1])
	if err != nil {
		s.logger.Error(err)
	}
	milliseconds /= 100

	return int64((hours*3600+minutes*60+seconds)*10 + milliseconds)
}

func (s *MediaServiceImpl) convertTimeToString(timeNum int64) string {
	milliseconds := (timeNum % 10) * 100
	seconds := timeNum / 10 % 60
	minutes := timeNum / 10 / 60 % 60
	hours := timeNum / 10 / 60 / 60

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

func (s *MediaServiceImpl) convertDurationToInt64(duration time.Duration) int64 {
	return duration.Milliseconds() / 100
}

func (s *MediaServiceImpl) GetAudioDurationWav(audioPath string) (time.Duration, int64, error) {
	var duration time.Duration
	file, err := os.Open(s.pathForMedia + audioPath)
	s.logger.Info(s.pathForMedia + audioPath)
	if err != nil {
		s.logger.Error(err)
		return duration, 0, err
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	duration, err = decoder.Duration()
	s.logger.Info(duration)
	durationInt := duration.Milliseconds() / int64(100)
	s.logger.Info(durationInt)
	return duration, durationInt, nil
}

func (s *MediaServiceImpl) GetAudioDurationMp3(audioPath string) (time.Duration, int64, error) {
	var duration time.Duration
	file, err := os.Open(s.pathForMedia + audioPath)
	if err != nil {
		s.logger.Error(err)
		return duration, 0, err
	}
	defer file.Close()

	var t int64

	d := mp3.NewDecoder(file)
	var f mp3.Frame
	skipped := 0

	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			s.logger.Error(err)
			return duration, 0, err
		}

		duration += f.Duration()
	}

	s.logger.Info(t)
	return duration, duration.Milliseconds() / 100, nil
}
