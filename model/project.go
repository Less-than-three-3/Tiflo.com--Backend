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
	ProjectId uuid.UUID `json:"projectId"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	UserId    uuid.UUID `json:"userId"`
}

type User struct {
	UserId   uuid.UUID `json:"userId"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
}
