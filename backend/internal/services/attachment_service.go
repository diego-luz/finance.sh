package services

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/finance-sh/finance-sh/internal/config"
	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/entities"
	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/storage"
	"github.com/google/uuid"
)

// ErrUnsupportedType is returned when an uploaded file's content type is not in
// the allowlist. Handlers map it to HTTP 422.
var ErrUnsupportedType = errors.New("tipo de arquivo não suportado")

// ErrFileTooLarge is returned when an upload exceeds ATTACHMENT_MAX_MB. Handlers
// map it to HTTP 413.
var ErrFileTooLarge = errors.New("arquivo excede o tamanho máximo permitido")

// allowedContentTypes is the set of accepted attachment content types.
var allowedContentTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"image/gif":       {},
	"application/pdf": {},
}

// AttachmentService handles receipt/comprovante uploads for transactions: it
// validates ownership and the file, persists the metadata row and stores the
// bytes in Postgres BYTEA (see pkg/storage). Everything is org-scoped.
type AttachmentService struct {
	attachments *repositories.AttachmentRepository
	txs         *repositories.TransactionRepository
	storage     *storage.Storage
	cfg         *config.Config
}

func NewAttachmentService(
	attachments *repositories.AttachmentRepository,
	txs *repositories.TransactionRepository,
	st *storage.Storage,
	cfg *config.Config,
) *AttachmentService {
	return &AttachmentService{attachments: attachments, txs: txs, storage: st, cfg: cfg}
}

// Upload validates the transaction (must belong to org), the content type and
// the size, then creates the metadata row first and streams the bytes into the
// row's BYTEA column. The two-step order (row -> blob) keeps storage and DB in
// sync: the row identifies the blob and is rolled back when the blob write
// fails.
func (s *AttachmentService) Upload(orgID, userID, txID uuid.UUID, fileName, contentType string, size int64, r io.Reader) (*dto.AttachmentDTO, error) {
	// Ownership: the transaction must belong to the caller's org (404 otherwise).
	if _, err := s.txs.FindByID(orgID, txID); err != nil {
		return nil, err
	}

	// Content-type allowlist.
	contentType = normalizeContentType(contentType)
	if _, ok := allowedContentTypes[contentType]; !ok {
		return nil, ErrUnsupportedType
	}

	// Size limit (ATTACHMENT_MAX_MB). size <= 0 means the caller could not report
	// it; we reject only when it is known to exceed the cap (the handler also caps
	// the read with MaxBytesReader as a hard backstop).
	maxBytes := int64(s.cfg.Storage.MaxAttachmentMB) * 1024 * 1024
	if size > maxBytes {
		return nil, ErrFileTooLarge
	}

	safeName := sanitizeFileName(fileName)

	// Create the metadata row first so the blob write has a target id. The
	// object_key column is unused (nullable) and left empty.
	a := &entities.Attachment{
		OrganizationID: orgID,
		TransactionID:  &txID,
		FileName:       safeName,
		ContentType:    contentType,
		Size:           size,
		UploadedBy:     userID,
	}
	if err := s.attachments.Create(a); err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := s.storage.Put(ctx, a.ID.String(), r, size, contentType); err != nil {
		// Roll the metadata row back so we never expose an attachment with no
		// bytes behind it.
		_ = s.attachments.Delete(orgID, a.ID)
		return nil, err
	}

	d := attachmentDTO(a)
	return &d, nil
}

// List returns the attachments of a transaction (org-scoped).
func (s *AttachmentService) List(orgID, txID uuid.UUID) ([]dto.AttachmentDTO, error) {
	rows, err := s.attachments.ListByTransaction(orgID, txID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AttachmentDTO, 0, len(rows))
	for i := range rows {
		out = append(out, attachmentDTO(&rows[i]))
	}
	return out, nil
}

// Download verifies org ownership and opens a stream of the BYTEA blob. The
// returned ReadCloser must be closed by the caller.
func (s *AttachmentService) Download(orgID, id uuid.UUID) (*dto.AttachmentDTO, io.ReadCloser, error) {
	a, err := s.attachments.FindByID(orgID, id)
	if err != nil {
		return nil, nil, err
	}
	rc, _, _, err := s.storage.Get(context.Background(), a.ID.String())
	if err != nil {
		return nil, nil, err
	}
	d := attachmentDTO(a)
	return &d, rc, nil
}

// Delete verifies org ownership, clears the blob (best-effort) and removes the
// row. The blob clear is best-effort because a soft-delete on the row is the
// authoritative removal: the next vacuum will reclaim TOAST pages.
func (s *AttachmentService) Delete(orgID, id uuid.UUID) error {
	a, err := s.attachments.FindByID(orgID, id)
	if err != nil {
		return err
	}
	_ = s.storage.Delete(context.Background(), a.ID.String())
	return s.attachments.Delete(orgID, id)
}

// normalizeContentType strips any parameters (e.g. "; charset=...") and lowercases.
func normalizeContentType(ct string) string {
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.ToLower(strings.TrimSpace(ct))
}

// sanitizeFileName strips any path component and dangerous characters, returning
// a safe base name. Falls back to "arquivo" when the result is empty.
func sanitizeFileName(name string) string {
	// Drop directory components from either separator style.
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "arquivo"
	}
	return name
}

func attachmentDTO(a *entities.Attachment) dto.AttachmentDTO {
	d := dto.AttachmentDTO{
		ID:          a.ID.String(),
		FileName:    a.FileName,
		ContentType: a.ContentType,
		Size:        a.Size,
		CreatedAt:   a.CreatedAt,
	}
	if a.TransactionID != nil {
		d.TransactionID = a.TransactionID.String()
	}
	return d
}
