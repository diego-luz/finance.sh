package services

import (
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

type TagService struct {
	tags *repositories.TagRepository
}

func NewTagService(tags *repositories.TagRepository) *TagService {
	return &TagService{tags: tags}
}

func (s *TagService) List(orgID uuid.UUID) ([]dto.TagDTO, error) {
	rows, err := s.tags.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TagDTO, 0, len(rows))
	for i := range rows {
		out = append(out, tagDTO(&rows[i]))
	}
	return out, nil
}

func (s *TagService) Create(orgID uuid.UUID, req dto.TagRequest) (*dto.TagDTO, error) {
	t := &entities.Tag{
		OrganizationID: orgID,
		Name:           req.Name,
		Color:          defaultStr(req.Color, "#6b7280"),
	}
	if err := s.tags.Create(t); err != nil {
		return nil, err
	}
	d := tagDTO(t)
	return &d, nil
}

func (s *TagService) Update(orgID, id uuid.UUID, req dto.TagRequest) (*dto.TagDTO, error) {
	t, err := s.tags.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	t.Name = req.Name
	if req.Color != "" {
		t.Color = req.Color
	}
	if err := s.tags.Update(t); err != nil {
		return nil, err
	}
	d := tagDTO(t)
	return &d, nil
}

func (s *TagService) Delete(orgID, id uuid.UUID) error {
	return s.tags.Delete(orgID, id)
}

func tagDTO(t *entities.Tag) dto.TagDTO {
	return dto.TagDTO{
		ID:    t.ID.String(),
		Name:  t.Name,
		Color: t.Color,
	}
}
