package model

import (
	"github.com/google/uuid"
	"time"
)

type AudioPart struct {
	PartId    uuid.UUID `json:"partId"`
	ProjectId uuid.UUID `json:"projectId"`
	Start     int64     `json:"start"`
	Duration  int64     `json:"duration"`
	Text      string    `json:"text"`
	Path      string    `json:"path"`
}

type Project struct {
	ProjectId  uuid.UUID   `json:"projectId" binding:"required"`
	Created    time.Time   `json:"created"`
	Name       string      `json:"name" binding:"required"`
	VideoPath  string      `json:"path" binding:"required"`
	AudioPath  string      `json:"-"`
	UserId     uuid.UUID   `json:"userId" binding:"required"`
	AudioParts []AudioPart `json:"audioParts" binding:"omitempty"`
}

type VoiceText struct {
	Text string `json:"text"`
}

type Image struct {
	Name string `json:"name"`
}
