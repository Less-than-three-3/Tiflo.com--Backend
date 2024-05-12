package ffmpeg

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
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
	splitPoint := s.ConvertTimeFromString(splitPointStr)
	firstPartEnd := splitPoint - start

	s.logger.Info("-ss ", "00:00:00.000", " -t ", s.convertTimeToString(firstPartEnd), s.pathForMedia+firstPartName.String()+".wav")

	_, err := exec.Command("ffmpeg", "-i", s.pathForMedia+audioPartToSplit.Path, "-vn", "-acodec", "pcm_s16le",
		"-ss", "00:00:00.000", "-t", s.convertTimeToString(firstPartEnd),
		s.pathForMedia+firstPartName.String()+".wav").Output()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	result = append(result, model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: audioPartToSplit.ProjectId,
		Start:     audioPartToSplit.Start,
		Duration:  splitPoint - audioPartToSplit.Start,
		Text:      "",
		Path:      firstPartName.String() + ".wav",
	})

	secondPartName := uuid.New()
	s.logger.Info("-ss ", s.convertTimeToString(firstPartEnd), " -t ", s.convertTimeToString(start+audioPartToSplit.Duration-splitPoint),
		s.pathForMedia+secondPartName.String()+".wav")

	_, err = exec.Command("ffmpeg", "-i", s.pathForMedia+audioPartToSplit.Path, "-vn", "-acodec", "pcm_s16le",
		"-ss", s.convertTimeToString(firstPartEnd), "-t", s.convertTimeToString(start+audioPartToSplit.Duration-splitPoint),
		s.pathForMedia+secondPartName.String()+".wav").Output()
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}

	result = append(result, model.AudioPart{
		PartId:    uuid.New(),
		ProjectId: audioPartToSplit.ProjectId,
		Start:     splitPoint + s.convertDurationToInt64(duration),
		Duration:  start + audioPartToSplit.Duration - splitPoint,
		Text:      "",
		Path:      secondPartName.String() + ".wav",
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

		if err = d.Decode(&f, &skipped); err != nil {
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

// ConcatAudio ffmpeg -i audio1.wav -i audio2.wav -i audio3.wav -i audio4.wav -i audio5.wav \
// -filter_complex '[0:0][1:0][2:0][3:0][4:0]concat=n=5:v=0:a=1[out]' \
// -map '[out]' output.wav
func (s *MediaServiceImpl) ConcatAudio(audioParts []model.AudioPart) (string, error) {
	sort.SliceStable(audioParts, func(i, j int) bool {
		return audioParts[i].Start < audioParts[j].Start
	})

	var arguments []string
	var filter = ""

	for i, part := range audioParts {
		arguments = append(arguments, "-i", s.pathForMedia+part.Path)
		filter += fmt.Sprintf("[%d:a]", i)
	}

	filter += fmt.Sprintf("concat=n=%d:v=0:a=1[out]", len(audioParts))

	s.logger.Info(filter)
	s.logger.Info(arguments)
	concatAudio := uuid.New()

	arguments = append(arguments, "-filter_complex", filter, "-map", "[out]", s.pathForMedia+concatAudio.String()+".wav")

	s.logger.Info(arguments)
	//
	//_, err := exec.Command("ffmpeg", arguments...).Output()
	//if err != nil {
	//	s.logger.Error(err)
	//	return "", err
	//}

	cmd := exec.Command("ffmpeg", arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("FFmpeg command failed: %v", err)
	}

	return concatAudio.String() + ".wav", nil
}
