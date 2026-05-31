package repositories

import (
	"errors"
	"strings"

	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrTagExists is returned when creating/updating a tag would collide with an
// existing tag of the same name in the organization (unique idx_org_tag_name).
// Handlers map it to HTTP 409.
var ErrTagExists = errors.New("já existe uma etiqueta com este nome")

// TagRepository persists tags (rótulos), always scoped by organization_id.
type TagRepository struct{ db *gorm.DB }

func NewTagRepository(db *gorm.DB) *TagRepository { return &TagRepository{db: db} }

func (r *TagRepository) Create(t *entities.Tag) error {
	if err := r.db.Create(t).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrTagExists
		}
		return err
	}
	return nil
}

func (r *TagRepository) Update(t *entities.Tag) error {
	if err := r.db.Save(t).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrTagExists
		}
		return err
	}
	return nil
}

// Delete removes a tag (org-scoped) and clears its rows from the join table so
// no dangling transaction_tags reference the deleted tag.
func (r *TagRepository) Delete(orgID, id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clear the association rows for this tag first (join has no soft delete).
		if err := tx.Exec("DELETE FROM transaction_tags WHERE tag_id = ?", id).Error; err != nil {
			return err
		}
		res := tx.Where("organization_id = ? AND id = ?", orgID, id).Delete(&entities.Tag{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *TagRepository) FindByID(orgID, id uuid.UUID) (*entities.Tag, error) {
	var t entities.Tag
	err := r.db.Where("organization_id = ? AND id = ?", orgID, id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, err
}

func (r *TagRepository) List(orgID uuid.UUID) ([]entities.Tag, error) {
	var tags []entities.Tag
	err := r.db.Where("organization_id = ?", orgID).Order("name asc").Find(&tags).Error
	return tags, err
}

// FindByIDs resolves the given ids to the org's tags, skipping any id that does
// not belong to the organization (tenant guard). Returns the matched tags.
func (r *TagRepository) FindByIDs(orgID uuid.UUID, ids []uuid.UUID) ([]entities.Tag, error) {
	if len(ids) == 0 {
		return []entities.Tag{}, nil
	}
	var tags []entities.Tag
	err := r.db.Where("organization_id = ? AND id IN ?", orgID, ids).Find(&tags).Error
	return tags, err
}

// SearchByName returns the org's tags whose name matches the query (ILIKE),
// ordered by name, capped at limit. Used by global search.
func (r *TagRepository) SearchByName(orgID uuid.UUID, q string, limit int) ([]entities.Tag, error) {
	var tags []entities.Tag
	err := r.db.Where("organization_id = ? AND name ILIKE ?", orgID, "%"+q+"%").
		Order("name asc").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

// isUniqueViolation reports whether the error is a Postgres unique-constraint
// violation. GORM wraps gorm.ErrDuplicatedKey for known dialects; the string
// fallback covers drivers that do not translate the SQLSTATE 23505.
func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "23505")
}
