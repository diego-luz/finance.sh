package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/cache"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

// normalizeText lowercases and strips diacritics (NFD + drop combining marks) so
// rule matching is accent-insensitive — Brazilian descriptions/keywords are often
// typed without accents ("combustivel" must match the rule "combustível").
func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ErrInvalidRule is returned when a categorization rule is malformed (e.g. a
// regex MatchType whose pattern does not compile). Handlers map it to HTTP 422.
var ErrInvalidRule = errors.New("regra de categorização inválida")

// stopwords are common pt-BR/short tokens dropped before the history fallback so
// the ILIKE search keys on meaningful words (e.g. merchant names) only.
var stopwords = map[string]bool{
	"com": true, "para": true, "por": true, "dos": true, "das": true,
	"uma": true, "ltda": true, "ltd": true, "sao": true, "são": true,
	"the": true, "and": true, "pix": true, "ted": true, "doc": true,
	"pagamento": true, "pagto": true, "compra": true, "debito": true,
	"débito": true, "credito": true, "crédito": true, "cartao": true,
	"cartão": true, "transferencia": true, "transferência": true,
}

// tokenSplit splits on any run of non-alphanumeric characters.
var tokenSplit = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// CategorizationService resolves a category for a transaction description, first
// by org-defined rules (highest priority wins) and then by a history fallback
// (the most-used category among past transactions matching a significant token).
// It also exposes CRUD over the rules.
type CategorizationService struct {
	rules      *repositories.CategoryRuleRepository
	categories *repositories.CategoryRepository
	txs        *repositories.TransactionRepository
	cache      *cache.Cache
}

func NewCategorizationService(
	rules *repositories.CategoryRuleRepository,
	categories *repositories.CategoryRepository,
	txs *repositories.TransactionRepository,
	c *cache.Cache,
) *CategorizationService {
	return &CategorizationService{rules: rules, categories: categories, txs: txs, cache: c}
}

func (s *CategorizationService) invalidateDashboard(orgID uuid.UUID) {
	s.cache.Delete(context.Background(), DashboardCacheKey(orgID))
}

// ----- Rule CRUD -----

func (s *CategorizationService) ListRules(orgID uuid.UUID) ([]dto.CategoryRuleDTO, error) {
	rules, err := s.rules.List(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CategoryRuleDTO, 0, len(rules))
	for i := range rules {
		out = append(out, categoryRuleDTO(&rules[i]))
	}
	return out, nil
}

func (s *CategorizationService) CreateRule(orgID uuid.UUID, req dto.CategoryRuleRequest) (*dto.CategoryRuleDTO, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, ErrInvalidRule
	}
	if _, err := s.categories.FindByID(orgID, categoryID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	matchType := normalizeMatchType(req.MatchType)
	if err := validateRulePattern(matchType, req.Pattern); err != nil {
		return nil, err
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	rule := &entities.CategoryRule{
		OrganizationID: orgID,
		Pattern:        req.Pattern,
		MatchType:      matchType,
		CategoryID:     categoryID,
		Priority:       req.Priority,
		Active:         active,
	}
	if err := s.rules.Create(rule); err != nil {
		return nil, err
	}
	// Reload so the Category association is populated in the response.
	saved, err := s.rules.FindByID(orgID, rule.ID)
	if err != nil {
		d := categoryRuleDTO(rule)
		return &d, nil
	}
	d := categoryRuleDTO(saved)
	return &d, nil
}

func (s *CategorizationService) UpdateRule(orgID, id uuid.UUID, req dto.CategoryRuleRequest) (*dto.CategoryRuleDTO, error) {
	rule, err := s.rules.FindByID(orgID, id)
	if err != nil {
		return nil, err
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, ErrInvalidRule
	}
	if _, err := s.categories.FindByID(orgID, categoryID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	matchType := normalizeMatchType(req.MatchType)
	if err := validateRulePattern(matchType, req.Pattern); err != nil {
		return nil, err
	}

	rule.Pattern = req.Pattern
	rule.MatchType = matchType
	rule.CategoryID = categoryID
	rule.Priority = req.Priority
	if req.Active != nil {
		rule.Active = *req.Active
	}
	if err := s.rules.Update(rule); err != nil {
		return nil, err
	}
	saved, err := s.rules.FindByID(orgID, rule.ID)
	if err != nil {
		d := categoryRuleDTO(rule)
		return &d, nil
	}
	d := categoryRuleDTO(saved)
	return &d, nil
}

func (s *CategorizationService) DeleteRule(orgID, id uuid.UUID) error {
	return s.rules.Delete(orgID, id)
}

// ----- Matching -----

// compiledRule is a rule with its regexp pre-compiled (when MatchType=regex) and
// its pattern lower-cased once, so a batch of descriptions can be matched cheaply.
type compiledRule struct {
	categoryID uuid.UUID
	kind       string // category kind: income|expense
	matchType  string
	patternLC  string
	re         *regexp.Regexp
}

// loadRules fetches active rules once and pre-compiles them; call it once per
// batch and pass the result to matchWith.
func (s *CategorizationService) loadRules(orgID uuid.UUID) ([]compiledRule, error) {
	rules, err := s.rules.ListActive(orgID)
	if err != nil {
		return nil, err
	}
	out := make([]compiledRule, 0, len(rules))
	for i := range rules {
		r := &rules[i]
		// A rule whose category was deleted (no preload) is skipped: it cannot
		// contribute a kind nor a valid assignment.
		if r.Category == nil {
			continue
		}
		cr := compiledRule{
			categoryID: r.CategoryID,
			kind:       string(r.Category.Kind),
			matchType:  r.MatchType,
			patternLC:  normalizeText(r.Pattern),
		}
		if r.MatchType == entities.MatchRegex {
			re, err := regexp.Compile("(?i)" + r.Pattern)
			if err != nil {
				// Skip an uncompilable rule rather than failing the whole batch.
				continue
			}
			cr.re = re
		}
		out = append(out, cr)
	}
	return out, nil
}

// matchWith tests pre-loaded rules (already ordered by priority desc) against a
// description, restricted to rules whose category kind == txType. Returns the
// first matching category id, or nil.
func matchWith(rules []compiledRule, description, txType string) *uuid.UUID {
	descLC := normalizeText(description)
	for i := range rules {
		r := &rules[i]
		if r.kind != txType {
			continue
		}
		switch r.matchType {
		case entities.MatchPrefix:
			if strings.HasPrefix(descLC, r.patternLC) {
				id := r.categoryID
				return &id
			}
		case entities.MatchRegex:
			if r.re != nil && r.re.MatchString(description) {
				id := r.categoryID
				return &id
			}
		default: // contains
			if r.patternLC != "" && strings.Contains(descLC, r.patternLC) {
				id := r.categoryID
				return &id
			}
		}
	}
	return nil
}

// Match resolves a category for one description: rules first (by priority), then
// the history fallback. Returns the category id and a source of
// "rule" | "history" | "none". Convenience wrapper that loads rules per call;
// for batches prefer loadRules + matchWith + the history step directly.
func (s *CategorizationService) Match(orgID uuid.UUID, description, txType string) (*uuid.UUID, string) {
	rules, err := s.loadRules(orgID)
	if err == nil {
		if id := matchWith(rules, description, txType); id != nil {
			return id, "rule"
		}
	}
	if id := s.historyMatch(orgID, description, txType); id != nil {
		return id, "history"
	}
	return nil, "none"
}

// historyMatch runs the history fallback for a single description.
func (s *CategorizationService) historyMatch(orgID uuid.UUID, description, txType string) *uuid.UUID {
	tokens := significantTokens(description)
	if len(tokens) == 0 {
		return nil
	}
	id, err := s.txs.MostCommonCategoryForToken(orgID, txType, tokens)
	if err != nil {
		return nil
	}
	return id
}

// Suggest wraps Match and, when a category is found, returns its DTO too.
func (s *CategorizationService) Suggest(orgID uuid.UUID, description, txType string) (dto.SuggestResponse, error) {
	id, source := s.Match(orgID, description, txType)
	resp := dto.SuggestResponse{Source: source}
	if id == nil {
		return resp, nil
	}
	resp.CategoryID = id.String()
	cat, err := s.categories.FindByID(orgID, *id)
	if err == nil {
		d := categoryDTO(cat)
		resp.Category = &d
	}
	return resp, nil
}

// ApplyToUncategorized runs the categorizer over every transaction in the org
// with no category and assigns the matched category. Rules and the history
// fallback are both applied. Updates are grouped by resolved category so each
// category needs a single UPDATE. Returns the number of rows updated.
func (s *CategorizationService) ApplyToUncategorized(orgID uuid.UUID) (int, error) {
	rules, err := s.loadRules(orgID)
	if err != nil {
		return 0, err
	}
	pending, err := s.txs.UncategorizedIDs(orgID)
	if err != nil {
		return 0, err
	}
	if len(pending) == 0 {
		return 0, nil
	}

	// Group matched transaction ids by the category they resolved to.
	byCategory := make(map[uuid.UUID][]uuid.UUID)
	for i := range pending {
		t := &pending[i]
		txType := string(t.Type)
		// Transfers have no income/expense category; skip them.
		if txType != string(entities.TxIncome) && txType != string(entities.TxExpense) {
			continue
		}
		id := matchWith(rules, t.Description, txType)
		if id == nil {
			id = s.historyMatch(orgID, t.Description, txType)
		}
		if id == nil {
			continue
		}
		byCategory[*id] = append(byCategory[*id], t.ID)
	}

	updated := 0
	for catID, ids := range byCategory {
		n, err := s.txs.BulkCategorize(orgID, ids, catID)
		if err != nil {
			return updated, err
		}
		updated += int(n)
	}
	if updated > 0 {
		s.invalidateDashboard(orgID)
	}
	return updated, nil
}

// ----- helpers -----

// significantTokens splits a description into lower-cased tokens, dropping those
// shorter than 3 chars and common stopwords. Deduplicated, capped at a few
// tokens so the history ILIKE query stays bounded.
func significantTokens(description string) []string {
	parts := tokenSplit.Split(strings.ToLower(description), -1)
	seen := make(map[string]bool, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) < 3 || stopwords[p] || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func normalizeMatchType(mt string) string {
	switch mt {
	case entities.MatchPrefix, entities.MatchRegex:
		return mt
	default:
		return entities.MatchContains
	}
}

// validateRulePattern rejects empty patterns and, for regex rules, patterns that
// do not compile.
func validateRulePattern(matchType, pattern string) error {
	if strings.TrimSpace(pattern) == "" {
		return ErrInvalidRule
	}
	if matchType == entities.MatchRegex {
		if _, err := regexp.Compile(pattern); err != nil {
			return ErrInvalidRule
		}
	}
	return nil
}

func categoryRuleDTO(r *entities.CategoryRule) dto.CategoryRuleDTO {
	d := dto.CategoryRuleDTO{
		ID:         r.ID.String(),
		Pattern:    r.Pattern,
		MatchType:  r.MatchType,
		CategoryID: r.CategoryID.String(),
		Priority:   r.Priority,
		Active:     r.Active,
	}
	if r.Category != nil {
		c := categoryDTO(r.Category)
		d.Category = &c
	}
	return d
}
