package client

import (
	"context"
	"fmt"

	"tiflo/pkg/grpc/generated"
	pb "tiflo/pkg/grpc/generated"

	"github.com/sirupsen/logrus"
)

type PythonClient struct {
	logger           *logrus.Entry
	voice2textClient generated.AIServiceClient
	image2textClient generated.ImageCaptioningClient
}

type AI interface {
	VoiceTheText(context context.Context, text string) (string, error)
	ImageToText(context context.Context, path string) (string, error)
}

func NewPythonClient(logger *logrus.Logger, voice2textClient generated.AIServiceClient, image2textClient generated.ImageCaptioningClient) *PythonClient {
	return &PythonClient{
		logger:           logger.WithField("component", "python_client"),
		voice2textClient: voice2textClient,
		image2textClient: image2textClient,
	}
}

func (p *PythonClient) VoiceTheText(context context.Context, text string) (string, error) {
	//p.logger.Info("text: ", text)
	fmt.Println("text: ", text)
	request := pb.TextToVoice{
		Text: text,
	}
	resp, err := p.voice2textClient.VoiceTheText(context, &request)
	if err != nil {
		p.logger.Error("voice the text: ", err)
		return "", err
	}

	p.logger.Info("answer", resp.Audio)
	return resp.Audio, nil
}

func (p *PythonClient) ImageToText(context context.Context, path string) (string, error) {
	p.logger.Info("path: ", path)

	request := pb.Image{
		ImagePath: path,
	}

	resp, err := p.image2textClient.ImageCaption(context, &request)
	if err != nil {
		p.logger.Error("image to text: ", err)
		return "", err
	}

	p.logger.Info("answer", resp.Text)
	return resp.Text, nil
}
