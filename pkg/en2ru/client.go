package en2ru

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"tiflo/model"

	"github.com/sirupsen/logrus"
)

type EnToRu interface {
	Translate(context context.Context, text string) (string, error)
}

type EnToRuClient struct {
	logger *logrus.Entry
	url    string
}

func NewEnToRuClient(logger *logrus.Logger, url string) *EnToRuClient {
	return &EnToRuClient{
		logger: logger.WithField("component", "EnToRuClient"),
		url:    url,
	}
}

func (c *EnToRuClient) Translate(context context.Context, text string) (string, error) {
	encodedText := url.QueryEscape(text)
	
	fullURL := fmt.Sprintf("%s?q=%s&langpair=en|ru", c.url, encodedText)
	
	resp, err := http.Get(fullURL)
	if err != nil {
		c.logger.Fatal("ошибка запроса: ", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Fatal("HTTP ошибка: ", resp.Status)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Fatal("ошибка чтения ответа: ", err)
		return "", err
	}

	var translationResponse model.TranslationResponse
	if err := json.Unmarshal(body, &translationResponse); err != nil {
		c.logger.Fatal("ошибка парсинга JSON: ", err)
		return "", err
	}

	if translationResponse.QuotaFinished {
		c.logger.Fatal("лимит запросов исчерпан")
		return "", err
	}

	return translationResponse.ResponseData.TranslatedText, nil
}
