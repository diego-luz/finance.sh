package services

import (
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/google/uuid"
)

type ContactService struct {
	contacts *repositories.ContactRepository
}

func NewContactService(contacts *repositories.ContactRepository) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) List(orgID uuid.UUID) ([]dto.ContactDTO, error) {
	rows, err := s.contacts.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ContactDTO, 0, len(rows))
	for i := range rows {
		out = append(out, contactDTO(&rows[i]))
	}
	return out, nil
}

func (s *ContactService) Get(orgID, id uuid.UUID) (*dto.ContactDTO, error) {
	c, err := s.contacts.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	d := contactDTO(c)
	return &d, nil
}

func (s *ContactService) Create(orgID uuid.UUID, req dto.ContactRequest) (*dto.ContactDTO, error) {
	c := &entities.Contact{
		OrganizationID: orgID,
		Name:           req.Name,
		Type:           contactType(req.Type),
		Document:       req.Document,
		Email:          req.Email,
		Phone:          req.Phone,
		Notes:          req.Notes,
	}
	if err := s.contacts.Create(c); err != nil {
		return nil, err
	}
	d := contactDTO(c)
	return &d, nil
}

func (s *ContactService) Update(orgID, id uuid.UUID, req dto.ContactRequest) (*dto.ContactDTO, error) {
	c, err := s.contacts.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	c.Name = req.Name
	c.Type = contactType(req.Type)
	c.Document = req.Document
	c.Email = req.Email
	c.Phone = req.Phone
	c.Notes = req.Notes
	if err := s.contacts.Update(c); err != nil {
		return nil, err
	}
	d := contactDTO(c)
	return &d, nil
}

func (s *ContactService) Delete(orgID, id uuid.UUID) error {
	return s.contacts.Delete(orgID, id)
}

// contactType defaults an empty/invalid type to "both".
func contactType(v string) entities.ContactType {
	switch entities.ContactType(v) {
	case entities.ContactCustomer, entities.ContactSupplier, entities.ContactBoth:
		return entities.ContactType(v)
	default:
		return entities.ContactBoth
	}
}

func contactDTO(c *entities.Contact) dto.ContactDTO {
	return dto.ContactDTO{
		ID:       c.ID.String(),
		Name:     c.Name,
		Type:     string(c.Type),
		Document: c.Document,
		Email:    c.Email,
		Phone:    c.Phone,
		Notes:    c.Notes,
	}
}
