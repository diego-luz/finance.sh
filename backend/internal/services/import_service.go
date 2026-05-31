package services

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/finance-sh/finance-sh/pkg/imports"
	"github.com/google/uuid"
)

// ErrEmptyImport is returned when a statement parses to zero usable rows.
// Handlers map it to HTTP 422.
var ErrEmptyImport = errors.New("nenhum lançamento encontrado no arquivo")

// ErrUnsupportedFormat is returned when the import format is not ofx/csv/auto
// or auto-detection fails. Handlers map it to HTTP 422.
var ErrUnsupportedFormat = errors.New("formato de arquivo não suportado")

// ImportService parses bank statements (OFX/CSV) and turns them into
// transactions in a stateless two-step flow: Preview parses + flags duplicates
// without persisting; Commit writes the reviewed rows in one DB transaction.
type ImportService struct {
	txs        *repositories.TransactionRepository
	accounts   *repositories.AccountRepository
	categories *repositories.CategoryRepository
	// categorization resolves a category from a row's description (rules +
	// history). import_service depends on categorization (one-way; categorization
	// only depends on repositories, so there is no import cycle).
	categorization *CategorizationService
	cache          *cache.Cache
}

func NewImportService(
	txs *repositories.TransactionRepository,
	accounts *repositories.AccountRepository,
	categories *repositories.CategoryRepository,
	categorization *CategorizationService,
	c *cache.Cache,
) *ImportService {
	return &ImportService{txs: txs, accounts: accounts, categories: categories, categorization: categorization, cache: c}
}

func (s *ImportService) invalidateDashboard(orgID uuid.UUID) {
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

// Preview parses the uploaded statement against the given account and returns
// the parsed rows with duplicate flags plus a summary. Nothing is persisted.
// format is "ofx", "csv" or "auto" (sniffed). csvOpts tunes the CSV parser.
func (s *ImportService) Preview(orgID, accountID uuid.UUID, format string, file io.Reader, csvOpts imports.CSVOptions) (dto.ImportPreview, error) {
	if _, err := s.accounts.FindByID(orgID, accountID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return dto.ImportPreview{}, ErrAccountNotInOrg
		}
		return dto.ImportPreview{}, err
	}

	resolved, rows, err := parseStatement(format, file, csvOpts)
	if err != nil {
		return dto.ImportPreview{}, err
	}
	if len(rows) == 0 {
		return dto.ImportPreview{}, ErrEmptyImport
	}

	// Batch-check existing external ids for the org+account (single query).
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.ExternalID != "" {
			ids = append(ids, r.ExternalID)
		}
	}
	existing, err := s.txs.ExistsExternalIDs(orgID, accountID, ids)
	if err != nil {
		return dto.ImportPreview{}, err
	}

	preview := dto.ImportPreview{Format: resolved, Rows: make([]dto.ImportRowDTO, 0, len(rows))}
	// In-file dedup: a statement may repeat the same row; flag the later one.
	seen := make(map[string]bool, len(rows))
	for i, r := range rows {
		dupe, reason := s.classifyRow(orgID, accountID, r, existing, seen)
		// Suggest a category for the row (rules + history). Best-effort: a failure
		// just leaves the suggestion empty.
		var suggested string
		if s.categorization != nil {
			if id, _ := s.categorization.Match(orgID, r.Description, r.Type); id != nil {
				suggested = id.String()
			}
		}
		preview.Rows = append(preview.Rows, dto.ImportRowDTO{
			Index:               i,
			Date:                r.Date,
			Description:         r.Description,
			Amount:              r.AmountCents,
			Type:                r.Type,
			ExternalID:          r.ExternalID,
			Duplicate:           dupe,
			Reason:              reason,
			SuggestedCategoryID: suggested,
		})
		if dupe {
			preview.Summary.Duplicates++
		} else {
			preview.Summary.New++
		}
	}
	preview.Summary.Total = len(rows)
	return preview, nil
}

// classifyRow decides whether a parsed row is a duplicate. Preference order:
// (1) an existing transaction with the same ExternalID (org+account); (2) the
// same ExternalID already seen earlier in this file; (3) when there is NO
// external id, an existing transaction with the same date+amount+description.
func (s *ImportService) classifyRow(orgID, accountID uuid.UUID, r imports.ParsedRow, existing map[string]bool, seen map[string]bool) (bool, string) {
	if r.ExternalID != "" {
		if existing[r.ExternalID] {
			return true, "já importado anteriormente"
		}
		if seen[r.ExternalID] {
			return true, "duplicado no arquivo"
		}
		seen[r.ExternalID] = true
		return false, ""
	}
	// No external id: fall back to a signature match against the DB.
	dupe, err := s.txs.ExistsBySignature(orgID, accountID, r.Date, r.AmountCents, r.Description)
	if err == nil && dupe {
		return true, "lançamento equivalente já existe"
	}
	return false, ""
}

// Commit creates the reviewed rows as transactions in the given account, in a
// single DB transaction. Duplicates (by ExternalID) are re-checked and skipped.
// An optional categoryID (validated against the org) is applied to every row.
// Returns the number of rows created and skipped.
func (s *ImportService) Commit(orgID, userID, accountID uuid.UUID, categoryID *uuid.UUID, rows []dto.ImportCommitRow) (created, skipped int, err error) {
	if len(rows) == 0 {
		return 0, 0, ErrEmptyImport
	}
	if _, err := s.accounts.FindByID(orgID, accountID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return 0, 0, ErrAccountNotInOrg
		}
		return 0, 0, err
	}

	if categoryID != nil {
		if _, err := s.categories.FindByID(orgID, *categoryID); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return 0, 0, ErrNotFound
			}
			return 0, 0, err
		}
	}

	// Re-check dedup for the rows that carry an external id (one query).
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.ExternalID != "" {
			ids = append(ids, r.ExternalID)
		}
	}
	existing, err := s.txs.ExistsExternalIDs(orgID, accountID, ids)
	if err != nil {
		return 0, 0, err
	}

	seen := make(map[string]bool, len(rows))
	toCreate := make([]*entities.Transaction, 0, len(rows))
	for _, r := range rows {
		// Validate the row defensively (the handler already validates via tags,
		// but Commit may be called with hand-built rows).
		if r.Amount <= 0 {
			skipped++
			continue
		}
		txType := entities.TransactionType(r.Type)
		if txType != entities.TxIncome && txType != entities.TxExpense {
			skipped++
			continue
		}
		if r.ExternalID != "" {
			if existing[r.ExternalID] || seen[r.ExternalID] {
				skipped++
				continue
			}
			seen[r.ExternalID] = true
		}

		// Category resolution: an explicit global categoryID always wins. When it
		// is absent, auto-assign via the categorizer (rules + history). Best-effort:
		// a non-match leaves the row uncategorized.
		rowCategory := categoryID
		if rowCategory == nil && s.categorization != nil {
			if id, _ := s.categorization.Match(orgID, r.Description, r.Type); id != nil {
				rowCategory = id
			}
		}

		t := &entities.Transaction{
			OrganizationID: orgID,
			AccountID:      accountID,
			CategoryID:     rowCategory,
			Type:           txType,
			Amount:         r.Amount,
			Description:    r.Description,
			Date:           r.Date,
			Paid:           true,
			ExternalID:     r.ExternalID,
		}
		toCreate = append(toCreate, t)
	}

	if len(toCreate) == 0 {
		return 0, skipped, nil
	}
	if err := s.txs.CreateMany(toCreate); err != nil {
		return 0, 0, err
	}
	s.invalidateDashboard(orgID)
	return len(toCreate), skipped, nil
}

// parseStatement resolves the format (sniffing when "auto") and dispatches to
// the matching parser. Returns the resolved format string and the parsed rows.
func parseStatement(format string, file io.Reader, csvOpts imports.CSVOptions) (string, []imports.ParsedRow, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "ofx":
		rows, err := imports.ParseOFX(file)
		return "ofx", rows, err
	case "csv":
		rows, err := imports.ParseCSV(file, csvOpts)
		return "csv", rows, err
	case "", "auto":
		// Sniff: buffer the head, decide, then parse the full content.
		data, err := io.ReadAll(file)
		if err != nil {
			return "", nil, err
		}
		if looksLikeOFX(data) {
			rows, perr := imports.ParseOFX(strings.NewReader(string(data)))
			return "ofx", rows, perr
		}
		rows, perr := imports.ParseCSV(strings.NewReader(string(data)), csvOpts)
		return "csv", rows, perr
	default:
		return "", nil, ErrUnsupportedFormat
	}
}

// looksLikeOFX returns true when the head of the file shows OFX markers
// ("OFXHEADER" for the SGML form, "<OFX" for the XML form).
func looksLikeOFX(data []byte) bool {
	head := strings.ToUpper(string(data))
	if len(head) > 4096 {
		head = head[:4096]
	}
	return strings.Contains(head, "OFXHEADER") || strings.Contains(head, "<OFX")
}
