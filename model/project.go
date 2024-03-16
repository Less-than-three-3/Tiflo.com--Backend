package model

import "github.com/google/uuid"

type ImageProject struct {
	ProjectId uuid.UUID `json:"projectId"`
	Image     string    `json:"image"`
	Name      string    `json:"name"`
	UserId    uuid.UUID
}
