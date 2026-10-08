package model

import "time"

type Workspace struct {
	ID string `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
