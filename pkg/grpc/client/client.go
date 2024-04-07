package client

import (
	"context"

	"tiflo/pkg/grpc/generated"
	pb "tiflo/pkg/grpc/generated"

	"github.com/sirupsen/logrus"
)

type PythonClient struct {
	logger *logrus.Entry
	client generated.AIServiceClient
}

type AI interface {
	VoiceTheText(context context.Context, text string) (string, error)
}

func NewPythonClient(logger *logrus.Logger, client generated.AIServiceClient) *PythonClient {
	return &PythonClient{
		logger: logger.WithField("component", "python_client"),
		client: client,
	}
}

func (p *PythonClient) VoiceTheText(context context.Context, text string) (string, error) {
	p.logger.Info("text: ", text)
	request := pb.TextToVoice{
		Text: text,
	}
	resp, err := p.client.VoiceTheText(context, &request)
	if err != nil {
		p.logger.Error("voice the text: ", err)
		return "", err
	}

	p.logger.Info("answer", resp.Audio)
	return resp.Audio, nil
}
