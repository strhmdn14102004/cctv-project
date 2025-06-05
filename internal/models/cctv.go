package models

import "time"

type CCTV struct {
	ID           int       `json:"id"`
	LocationID   int       `json:"-"`
	Location     *Location `json:"location,omitempty"`
	Name         string    `json:"name"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	SourceURL    string    `json:"sourceUrl"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateCCTVRequest struct {
	LocationID   int     `json:"locationId" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	ThumbnailURL *string `json:"thumbnailUrl"`
	SourceURL    string  `json:"sourceUrl" validate:"required,url"`
}

type UpdateCCTVRequest struct {
	LocationID   *int    `json:"locationId"`
	Name         *string `json:"name"`
	ThumbnailURL *string `json:"thumbnailUrl"`
	SourceURL    *string `json:"sourceUrl" validate:"omitempty,url"`
	IsActive     *bool   `json:"isActive"`
}
