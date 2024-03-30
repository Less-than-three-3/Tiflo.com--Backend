package model

import "github.com/google/uuid"

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
	Name       string      `json:"name" binding:"required"`
	Path       string      `json:"path" binding:"required"`
	UserId     uuid.UUID   `json:"userId" binding:"required"`
	AudioParts []AudioPart `json:"audioParts" binding:"omitempty"`
}
