// Package storage persists attachment blobs (receipts/comprovantes) directly
// into Postgres BYTEA. The deployment is single-host and the typical receipt
// fits well within Postgres TOAST limits (cap is 10MB), so an external object
// store is unnecessary. Callers talk to Storage.Put / Get / Delete with the
// attachment row id as the key.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"gorm.io/gorm"
)

// ErrStorageUnavailable is returned by every method when the underlying
// database handle is nil. Handlers map it to HTTP 503.
var ErrStorageUnavailable = errors.New("armazenamento de arquivos indisponível")

// ErrNotFound is returned by Get when no row matches the key (or the row was
// soft-deleted).
var ErrNotFound = errors.New("anexo não encontrado")

// Storage persists blobs in the `attachments.data` BYTEA column. The row is
// expected to exist before Put is called; Put updates the BYTEA + content
// metadata of the matching row.
type Storage struct {
	db  *gorm.DB
	log *slog.Logger
}

// New returns a Storage bound to db.
func New(db *gorm.DB, log *slog.Logger) (*Storage, error) {
	if log == nil {
		log = slog.Default()
	}
	if db == nil {
		return nil, ErrStorageUnavailable
	}
	return &Storage{db: db, log: log}, nil
}

// Put writes the blob into the attachments row identified by key (the row id
// as a string). The row must already exist; AttachmentService creates the row
// first and then calls Put.
func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s == nil || s.db == nil {
		return ErrStorageUnavailable
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("storage: read body: %w", err)
	}
	res := s.db.WithContext(ctx).Exec(
		`UPDATE attachments
		 SET data = ?, content_type = ?, size = ?, size_bytes = ?
		 WHERE id = ?`,
		buf, contentType, size, size, key,
	)
	if res.Error != nil {
		return fmt.Errorf("storage: update blob: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Get returns a reader over the BYTEA blob plus its size and content type.
// Caller must Close() the returned ReadCloser.
func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, int64, string, error) {
	if s == nil || s.db == nil {
		return nil, 0, "", ErrStorageUnavailable
	}
	var row struct {
		Data        []byte
		ContentType string
		Size        int64
	}
	err := s.db.WithContext(ctx).Raw(
		`SELECT data, content_type, size
		 FROM attachments
		 WHERE id = ? AND deleted_at IS NULL`,
		key,
	).Scan(&row).Error
	if err != nil {
		return nil, 0, "", fmt.Errorf("storage: read blob: %w", err)
	}
	if row.Data == nil {
		return nil, 0, "", ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(row.Data)), row.Size, row.ContentType, nil
}

// Delete clears the BYTEA blob on the row. The row itself is removed by the
// service layer; clearing the blob is safe to call before that and on
// already-cleared rows.
func (s *Storage) Delete(ctx context.Context, key string) error {
	if s == nil || s.db == nil {
		return ErrStorageUnavailable
	}
	return s.db.WithContext(ctx).Exec(
		`UPDATE attachments SET data = NULL WHERE id = ?`, key,
	).Error
}
