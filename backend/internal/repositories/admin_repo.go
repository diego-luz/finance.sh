package repositories

import (
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRepository backs the platform back-office (super-admin). It is NOT
// tenant-scoped: every read/write here is platform-wide. Listings use grouped
// count queries (resolved into maps) to avoid N+1 round-trips.
type AdminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) *AdminRepository { return &AdminRepository{db: db} }

// ----- Global stats -----

// AdminCounts holds the platform-wide aggregate counters for GET /admin/stats.
type AdminCounts struct {
	Organizations int64
	Users         int64
	Transactions  int64
}

// Stats computes the global counters in a handful of grouped queries.
func (r *AdminRepository) Stats() (AdminCounts, error) {
	var c AdminCounts
	if err := r.db.Model(&entities.Organization{}).Count(&c.Organizations).Error; err != nil {
		return c, err
	}
	if err := r.db.Model(&entities.User{}).Count(&c.Users).Error; err != nil {
		return c, err
	}
	if err := r.db.Model(&entities.Transaction{}).Count(&c.Transactions).Error; err != nil {
		return c, err
	}
	return c, nil
}

// ----- Organizations listing -----

// ListOrganizations returns a page of organizations (newest first) optionally
// filtered by a name/slug search, plus the total matching count.
func (r *AdminRepository) ListOrganizations(search string, page, perPage int) ([]entities.Organization, int64, error) {
	q := r.db.Model(&entities.Organization{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name ILIKE ? OR slug ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orgs []entities.Organization
	err := q.Order("created_at desc").
		Offset((page - 1) * perPage).Limit(perPage).
		Find(&orgs).Error
	return orgs, total, err
}

// OwnersByIDs batch-loads the owner users for the given ids into an id->user map.
func (r *AdminRepository) OwnersByIDs(ids []uuid.UUID) (map[uuid.UUID]entities.User, error) {
	out := make(map[uuid.UUID]entities.User, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var users []entities.User
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		out[users[i].ID] = users[i]
	}
	return out, nil
}

// MemberCountsByOrg returns a map org_id -> member count for the given orgs in a
// single grouped query.
func (r *AdminRepository) MemberCountsByOrg(orgIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	out := make(map[uuid.UUID]int64, len(orgIDs))
	if len(orgIDs) == 0 {
		return out, nil
	}
	type row struct {
		OrganizationID uuid.UUID
		N              int64
	}
	var rows []row
	err := r.db.Model(&entities.Membership{}).
		Select("organization_id, count(*) as n").
		Where("organization_id IN ?", orgIDs).
		Group("organization_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, x := range rows {
		out[x.OrganizationID] = x.N
	}
	return out, nil
}

// TransactionCountsByOrg returns a map org_id -> transaction count for the given
// orgs in a single grouped query.
func (r *AdminRepository) TransactionCountsByOrg(orgIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	out := make(map[uuid.UUID]int64, len(orgIDs))
	if len(orgIDs) == 0 {
		return out, nil
	}
	type row struct {
		OrganizationID uuid.UUID
		N              int64
	}
	var rows []row
	err := r.db.Model(&entities.Transaction{}).
		Select("organization_id, count(*) as n").
		Where("organization_id IN ?", orgIDs).
		Group("organization_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, x := range rows {
		out[x.OrganizationID] = x.N
	}
	return out, nil
}

// ----- Users listing -----

// ListUsers returns a page of users (newest first) optionally filtered by a
// name/email search, plus the total matching count.
func (r *AdminRepository) ListUsers(search string, page, perPage int) ([]entities.User, int64, error) {
	q := r.db.Model(&entities.User{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []entities.User
	err := q.Order("created_at desc").
		Offset((page - 1) * perPage).Limit(perPage).
		Find(&users).Error
	return users, total, err
}

// OrgNamesByUser returns a map user_id -> list of organization names the user
// belongs to, for the given user ids, in a single joined query.
func (r *AdminRepository) OrgNamesByUser(userIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	out := make(map[uuid.UUID][]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	type row struct {
		UserID uuid.UUID
		Name   string
	}
	var rows []row
	err := r.db.Table("memberships").
		Select("memberships.user_id as user_id, organizations.name as name").
		Joins("JOIN organizations ON organizations.id = memberships.organization_id").
		Where("memberships.user_id IN ? AND memberships.deleted_at IS NULL AND organizations.deleted_at IS NULL", userIDs).
		Order("organizations.name asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, x := range rows {
		out[x.UserID] = append(out[x.UserID], x.Name)
	}
	return out, nil
}
