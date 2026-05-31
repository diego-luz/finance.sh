package services

import (
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

type GoalService struct {
	goals *repositories.GoalRepository
}

func NewGoalService(goals *repositories.GoalRepository) *GoalService {
	return &GoalService{goals: goals}
}

func (s *GoalService) List(orgID uuid.UUID) ([]dto.GoalDTO, error) {
	goals, err := s.goals.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.GoalDTO, 0, len(goals))
	for i := range goals {
		out = append(out, goalDTO(&goals[i]))
	}
	return out, nil
}

func (s *GoalService) Get(orgID, id uuid.UUID) (*dto.GoalDTO, error) {
	g, err := s.goals.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := goalDTO(g)
	return &d, nil
}

func (s *GoalService) Create(orgID uuid.UUID, req dto.GoalRequest) (*dto.GoalDTO, error) {
	g := &entities.Goal{
		OrganizationID: orgID,
		Name:           req.Name,
		TargetAmount:   req.TargetAmount,
		CurrentAmount:  req.CurrentAmount,
		Deadline:       req.Deadline,
		Color:          defaultStr(req.Color, "#10b981"),
	}
	if err := s.goals.Create(g); err != nil {
		return nil, err
	}
	d := goalDTO(g)
	return &d, nil
}

func (s *GoalService) Update(orgID, id uuid.UUID, req dto.GoalRequest) (*dto.GoalDTO, error) {
	g, err := s.goals.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	g.Name = req.Name
	g.TargetAmount = req.TargetAmount
	g.CurrentAmount = req.CurrentAmount
	g.Deadline = req.Deadline
	if req.Color != "" {
		g.Color = req.Color
	}
	if err := s.goals.Update(g); err != nil {
		return nil, err
	}
	d := goalDTO(g)
	return &d, nil
}

func (s *GoalService) Delete(orgID, id uuid.UUID) error {
	return s.goals.Delete(orgID, id)
}

func goalDTO(g *entities.Goal) dto.GoalDTO {
	var progress float64
	if g.TargetAmount > 0 {
		progress = float64(g.CurrentAmount) / float64(g.TargetAmount)
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return dto.GoalDTO{
		ID:            g.ID.String(),
		Name:          g.Name,
		TargetAmount:  g.TargetAmount,
		CurrentAmount: g.CurrentAmount,
		Deadline:      g.Deadline,
		Color:         g.Color,
		Progress:      progress,
	}
}
