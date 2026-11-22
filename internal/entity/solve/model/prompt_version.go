package model

import "time"

type PromptVersion struct {
	Stage     string    `json:"stage"`
	Template  string    `json:"template"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}
