package model

import "time"

type Bid struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	TenderID    string       `json:"tenderId"`
	AuthorType  string       `json:"authorType"`
	AuthorID    string       `json:"authorId"`
	Status      string       `json:"status"`
	Decision    string       `json:"decision,omitempty"`
	Feedback    string       `json:"feedback,omitempty"`
	Version     int          `json:"version"`
	CreatedAt   time.Time    `json:"createdAt"`
	History     []BidVersion `json:"history,omitempty"`
}

type BidVersion struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Decision    string    `json:"decision,omitempty"`
	Feedback    string    `json:"feedback,omitempty"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

// BidReview represents review on a bid.
type BidReview struct {
	ID             string    `json:"id"`
	TenderID       string    `json:"tenderId"`
	AuthorUsername string    `json:"authorUsername"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"createdAt"`
}
