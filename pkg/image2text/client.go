package image2text

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ImageToTextPython interface {
	ImageToText(context context.Context, text string) (string, error)
}

type ImageToTextClient struct {
	logger *logrus.Entry
	host   string
}

func NewITTClient(logger *logrus.Logger, host string) *ImageToTextClient {
	return &ImageToTextClient{
		logger: logger.WithField("component", "ImageToTextClient"),
		host:   host,
	}
}

func (c *ImageToTextClient) ImageToText(context context.Context, filepath string) (string, error) {
	// return "Это какая то красивая картинка", nil

	c.logger.Info("Open")
	file, err := os.Open(filepath)
	if err != nil {
		c.logger.Fatal("Ошибка открытия файла:", err)
		return "", errors.New("Ошибка открытия файла:" + err.Error())
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	c.logger.Info("CreateFormFile")

	part, err := writer.CreateFormFile("image", filepath)
	if err != nil {
		c.logger.Fatal("Ошибка создания формы:", err)
		return "", errors.New("Ошибка создания формы:" + err.Error())
	}
	c.logger.Info("Copy")
	if _, err := io.Copy(part, file); err != nil {
		c.logger.Fatal("Ошибка копирования файла:", err)
		return "", errors.New("Ошибка копирования файла:" + err.Error())
	}
	writer.Close()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	c.logger.Info("NewRequest")
	req, err := http.NewRequest("POST", c.host, body)
	if err != nil {
		c.logger.Fatal("Ошибка создания запроса:", err)
		return "", errors.New("Ошибка создания запроса:" + err.Error())
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.logger.Info("Do")
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Fatal("Ошибка отправки запроса:", err)
		return "", errors.New("Ошибка отправки запроса:" + err.Error())
	}
	defer resp.Body.Close()
	c.logger.Info("resp")
	if resp.StatusCode != http.StatusOK {
		c.logger.Fatalf("Сервер вернул ошибку: %s", resp.Status)
		return "", errors.New("Сервер вернул ошибку:" + resp.Status)
	}

	var result struct {
		Description string `json:"description"`
		Error       string `json:"error,omitempty"`
	}
	c.logger.Info("NewDecoder")
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Fatal("Ошибка декодирования JSON:", err)
		return "", errors.New("Ошибка декодирования JSON:" + err.Error())
	}
	c.logger.Info("result")
	if result.Error != "" {
		c.logger.Fatal("Ошибка сервера:", result.Error)
		return "", errors.New("Ошибка сервера:" + result.Error)
	}
	c.logger.Info("Println")
	fmt.Println("Описание изображения:", result.Description)

	return result.Description, nil
}
