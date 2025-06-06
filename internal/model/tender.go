package model

import "time"

type Tender struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	ServiceType     string          `json:"serviceType"`
	OrganizationID  string          `json:"organizationId"`
	CreatorUsername string          `json:"creatorUsername"`
	Status          string          `json:"status"`
	Version         int             `json:"version"`
	CreatedAt       time.Time       `json:"createdAt"`
	History         []TenderVersion `json:"history,omitempty"`
}

type TenderVersion struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}
