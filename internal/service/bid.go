package service

import (
    "errors"
    "sort"
    "time"

    "github.com/google/uuid"

    "tender/internal/model"
    "tender/internal/storage"
)

type BidService struct {
    repo *storage.Storage
}

func NewBidService(r *storage.Storage) *BidService {
    return &BidService{repo: r}
}

func (s *BidService) Create(name, desc, tenderID, authorType, authorID string) (model.Bid, error) {
    b := model.Bid{
        ID:          uuid.New().String(),
        Name:        name,
        Description: desc,
        TenderID:    tenderID,
        AuthorType:  authorType,
        AuthorID:    authorID,
        Status:      "Created",
        Version:     1,
        CreatedAt:   time.Now().UTC(),
    }
    if err := s.repo.AddBid(b); err != nil {
        return model.Bid{}, err
    }
    return b, nil
}

func (s *BidService) UserBids(username string) []model.Bid {
    res := make([]model.Bid, 0)
    for _, b := range s.repo.Data.Bids {
        if b.AuthorID == username {
            res = append(res, b)
        }
    }
    sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
    return res
}

func (s *BidService) ListForTender(tenderID string) []model.Bid {
    res := make([]model.Bid, 0)
    for _, b := range s.repo.Data.Bids {
        if b.TenderID == tenderID {
            res = append(res, b)
        }
    }
    sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
    return res
}

func (s *BidService) UpdateStatus(id, status string) (model.Bid, error) {
    bid, ok := s.repo.GetBid(id)
    if !ok {
        return model.Bid{}, errors.New("not found")
    }
    bid.History = append(bid.History, model.BidVersion{
        Name:        bid.Name,
        Description: bid.Description,
        Status:      bid.Status,
        Decision:    bid.Decision,
        Feedback:    bid.Feedback,
        Version:     bid.Version,
        CreatedAt:   bid.CreatedAt,
    })
    bid.Status = status
    bid.Version++
    if err := s.repo.UpdateBid(bid); err != nil {
        return model.Bid{}, err
    }
    return bid, nil
}

func (s *BidService) Edit(id string, name, desc *string) (model.Bid, error) {
    bid, ok := s.repo.GetBid(id)
    if !ok {
        return model.Bid{}, errors.New("not found")
    }
    bid.History = append(bid.History, model.BidVersion{
        Name:        bid.Name,
        Description: bid.Description,
        Status:      bid.Status,
        Decision:    bid.Decision,
        Feedback:    bid.Feedback,
        Version:     bid.Version,
        CreatedAt:   bid.CreatedAt,
    })
    if name != nil {
        bid.Name = *name
    }
    if desc != nil {
        bid.Description = *desc
    }
    bid.Version++
    if err := s.repo.UpdateBid(bid); err != nil {
        return model.Bid{}, err
    }
    return bid, nil
}

func (s *BidService) Decision(id, decision string) (model.Bid, error) {
    bid, ok := s.repo.GetBid(id)
    if !ok {
        return model.Bid{}, errors.New("not found")
    }
    bid.Decision = decision
    if err := s.repo.UpdateBid(bid); err != nil {
        return model.Bid{}, err
    }
    return bid, nil
}

func (s *BidService) Feedback(id, feedback string) (model.Bid, error) {
    bid, ok := s.repo.GetBid(id)
    if !ok {
        return model.Bid{}, errors.New("not found")
    }
    bid.Feedback = feedback
    if err := s.repo.UpdateBid(bid); err != nil {
        return model.Bid{}, err
    }
    return bid, nil
}

func (s *BidService) Rollback(id string, ver int) (model.Bid, error) {
    bid, ok := s.repo.GetBid(id)
    if !ok {
        return model.Bid{}, errors.New("not found")
    }
    if ver < 1 || ver > len(bid.History) {
        return model.Bid{}, errors.New("version not found")
    }
    snap := bid.History[ver-1]
    bid.History = append(bid.History, model.BidVersion{
        Name:        bid.Name,
        Description: bid.Description,
        Status:      bid.Status,
        Decision:    bid.Decision,
        Feedback:    bid.Feedback,
        Version:     bid.Version,
        CreatedAt:   bid.CreatedAt,
    })
    bid.Name = snap.Name
    bid.Description = snap.Description
    bid.Status = snap.Status
    bid.Decision = snap.Decision
    bid.Feedback = snap.Feedback
    bid.Version++
    if err := s.repo.UpdateBid(bid); err != nil {
        return model.Bid{}, err
    }
    return bid, nil
}

func (s *BidService) Reviews(tenderID, author string) []model.BidReview {
    return s.repo.ListReviews(tenderID, author)
}

