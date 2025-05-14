package text2speech

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TextToSpeech interface {
	TextToSpeech(context context.Context, text string) (string, error)
}

type TextToSpeechClient struct {
	logger *logrus.Entry
	url    string
	apiKey string
}

func NewTTSClient(logger *logrus.Logger, url string, apiKey string) *TextToSpeechClient {
	return &TextToSpeechClient{
		logger: logger.WithField("component", "TextToSpeechClient"),
		url:    url,
		apiKey: apiKey,
	}
}

func (c *TextToSpeechClient) TextToSpeech(context context.Context, text string) (string, error) {
	body := []byte("text=" + text +
		"&lang=ru-RU" +
		"&voice=alena" +
		"&format=mp3" +
		"&sampleRateHertz=48000")

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Fatal("Error creating request: ", err)
		return "", err
	}

	req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Fatal("Request failed: ", err)
		return "", err
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Fatal("API error: ", resp.Status, body)
		return "", errors.New("API error")
	}

	// Создаем файл с расширением .mp3
	filename := uuid.New().String() + ".mp3"
	outFile, err := os.Create("/media/" + filename)
	if err != nil {
		c.logger.Fatal("File creation error: ", err)
		return "", err
	}
	defer outFile.Close()

	// Копируем аудио данные в файл
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		c.logger.Fatal("Error saving file: ", err)
		return "", err
	}

	return filename, nil
}
