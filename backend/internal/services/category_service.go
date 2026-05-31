package services

import (
	"context"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
)

type CategoryService struct {
	categories *repositories.CategoryRepository
	cache      *cache.Cache
}

func NewCategoryService(categories *repositories.CategoryRepository, c *cache.Cache) *CategoryService {
	return &CategoryService{categories: categories, cache: c}
}

func (s *CategoryService) invalidateDashboard(orgID uuid.UUID) {
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

func (s *CategoryService) List(orgID uuid.UUID) ([]dto.CategoryDTO, error) {
	cats, err := s.categories.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CategoryDTO, 0, len(cats))
	for i := range cats {
		out = append(out, categoryDTO(&cats[i]))
	}
	return out, nil
}

func (s *CategoryService) Get(orgID, id uuid.UUID) (*dto.CategoryDTO, error) {
	c, err := s.categories.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := categoryDTO(c)
	return &d, nil
}

func (s *CategoryService) Create(orgID uuid.UUID, req dto.CategoryRequest) (*dto.CategoryDTO, error) {
	c := &entities.Category{
		OrganizationID: orgID,
		Name:           req.Name,
		Kind:           entities.CategoryKind(req.Kind),
		Color:          defaultStr(req.Color, "#6366f1"),
		Icon:           defaultStr(req.Icon, "tag"),
	}
	if err := s.categories.Create(c); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	d := categoryDTO(c)
	return &d, nil
}

func (s *CategoryService) Update(orgID, id uuid.UUID, req dto.CategoryRequest) (*dto.CategoryDTO, error) {
	c, err := s.categories.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	c.Name = req.Name
	c.Kind = entities.CategoryKind(req.Kind)
	if req.Color != "" {
		c.Color = req.Color
	}
	if req.Icon != "" {
		c.Icon = req.Icon
	}
	if err := s.categories.Update(c); err != nil {
		return nil, err
	}
	s.invalidateDashboard(orgID)
	d := categoryDTO(c)
	return &d, nil
}

func (s *CategoryService) Delete(orgID, id uuid.UUID) error {
	if err := s.categories.Delete(orgID, id); err != nil {
		return err
	}
	s.invalidateDashboard(orgID)
	return nil
}

func categoryDTO(c *entities.Category) dto.CategoryDTO {
	return dto.CategoryDTO{
		ID:    c.ID.String(),
		Name:  c.Name,
		Kind:  string(c.Kind),
		Color: c.Color,
		Icon:  c.Icon,
	}
}
