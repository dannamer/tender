package service

import (
    "errors"
    "sort"
    "time"

    "github.com/google/uuid"

    "tender/internal/model"
    "tender/internal/storage"
)

type TenderService struct {
    repo *storage.Storage
}

func NewTenderService(r *storage.Storage) *TenderService {
    return &TenderService{repo: r}
}

func (s *TenderService) List(serviceTypes []string) []model.Tender {
    res := make([]model.Tender, 0)
    for _, t := range s.repo.Data.Tenders {
        if len(serviceTypes) == 0 || contains(serviceTypes, t.ServiceType) {
            res = append(res, t)
        }
    }
    sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
    return res
}

func (s *TenderService) Create(name, desc, serviceType, orgID, username string) (model.Tender, error) {
    t := model.Tender{
        ID:              uuid.New().String(),
        Name:            name,
        Description:     desc,
        ServiceType:     serviceType,
        OrganizationID:  orgID,
        CreatorUsername: username,
        Status:          "Created",
        Version:         1,
        CreatedAt:       time.Now().UTC(),
    }
    if err := s.repo.AddTender(t); err != nil {
        return model.Tender{}, err
    }
    return t, nil
}

func (s *TenderService) UserTenders(username string) []model.Tender {
    res := make([]model.Tender, 0)
    for _, t := range s.repo.Data.Tenders {
        if t.CreatorUsername == username {
            res = append(res, t)
        }
    }
    sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
    return res
}

func (s *TenderService) UpdateStatus(id, status string) (model.Tender, error) {
    tender, ok := s.repo.GetTender(id)
    if !ok {
        return model.Tender{}, errors.New("not found")
    }
    tender.History = append(tender.History, model.TenderVersion{
        Name:        tender.Name,
        Description: tender.Description,
        ServiceType: tender.ServiceType,
        Status:      tender.Status,
        Version:     tender.Version,
        CreatedAt:   tender.CreatedAt,
    })
    tender.Status = status
    tender.Version++
    if err := s.repo.UpdateTender(tender); err != nil {
        return model.Tender{}, err
    }
    return tender, nil
}

func (s *TenderService) Edit(id string, name, desc, serviceType *string) (model.Tender, error) {
    tender, ok := s.repo.GetTender(id)
    if !ok {
        return model.Tender{}, errors.New("not found")
    }
    tender.History = append(tender.History, model.TenderVersion{
        Name:        tender.Name,
        Description: tender.Description,
        ServiceType: tender.ServiceType,
        Status:      tender.Status,
        Version:     tender.Version,
        CreatedAt:   tender.CreatedAt,
    })
    if name != nil {
        tender.Name = *name
    }
    if desc != nil {
        tender.Description = *desc
    }
    if serviceType != nil {
        tender.ServiceType = *serviceType
    }
    tender.Version++
    if err := s.repo.UpdateTender(tender); err != nil {
        return model.Tender{}, err
    }
    return tender, nil
}

func (s *TenderService) Rollback(id string, ver int) (model.Tender, error) {
    tender, ok := s.repo.GetTender(id)
    if !ok {
        return model.Tender{}, errors.New("not found")
    }
    if ver < 1 || ver > len(tender.History) {
        return model.Tender{}, errors.New("version not found")
    }
    snap := tender.History[ver-1]
    tender.History = append(tender.History, model.TenderVersion{
        Name:        tender.Name,
        Description: tender.Description,
        ServiceType: tender.ServiceType,
        Status:      tender.Status,
        Version:     tender.Version,
        CreatedAt:   tender.CreatedAt,
    })
    tender.Name = snap.Name
    tender.Description = snap.Description
    tender.ServiceType = snap.ServiceType
    tender.Status = snap.Status
    tender.Version++
    if err := s.repo.UpdateTender(tender); err != nil {
        return model.Tender{}, err
    }
    return tender, nil
}

func contains(arr []string, s string) bool {
    for _, a := range arr {
        if a == s {
            return true
        }
    }
    return false
}

