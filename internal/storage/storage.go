package storage

import (
	"encoding/json"
	"os"
	"sync"

	"tender/internal/model"
)

type Data struct {
	Tenders []model.Tender    `json:"tenders"`
	Bids    []model.Bid       `json:"bids"`
	Reviews []model.BidReview `json:"reviews"`
}

type Storage struct {
	path string
	mu   sync.Mutex
	Data Data
}

func New(path string) (*Storage, error) {
	s := &Storage{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) load() error {
	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&s.Data)
}

func (s *Storage) save() error {
	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(s.Data)
}

func (s *Storage) AddTender(t model.Tender) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data.Tenders = append(s.Data.Tenders, t)
	return s.save()
}

func (s *Storage) UpdateTender(t model.Tender) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Data.Tenders {
		if s.Data.Tenders[i].ID == t.ID {
			s.Data.Tenders[i] = t
			return s.save()
		}
	}
	return os.ErrNotExist
}

func (s *Storage) GetTender(id string) (model.Tender, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.Data.Tenders {
		if t.ID == id {
			return t, true
		}
	}
	return model.Tender{}, false
}

func (s *Storage) AddBid(b model.Bid) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data.Bids = append(s.Data.Bids, b)
	return s.save()
}

func (s *Storage) UpdateBid(b model.Bid) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Data.Bids {
		if s.Data.Bids[i].ID == b.ID {
			s.Data.Bids[i] = b
			return s.save()
		}
	}
	return os.ErrNotExist
}

func (s *Storage) GetBid(id string) (model.Bid, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.Data.Bids {
		if b.ID == id {
			return b, true
		}
	}
	return model.Bid{}, false
}

func (s *Storage) AddReview(r model.BidReview) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data.Reviews = append(s.Data.Reviews, r)
	return s.save()
}

func (s *Storage) ListReviews(tenderID, author string) []model.BidReview {
	s.mu.Lock()
	defer s.mu.Unlock()
	var res []model.BidReview
	for _, r := range s.Data.Reviews {
		if r.TenderID == tenderID && r.AuthorUsername == author {
			res = append(res, r)
		}
	}
	return res
}
