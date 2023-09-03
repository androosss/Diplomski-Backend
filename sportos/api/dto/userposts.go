package dto

import "time"

type UserPost struct {
	Name       string    `json:"name,omitempty"`
	UserText   string    `json:"userText,omitempty"`
	ImageNames []string  `json:"imageNames,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
}
