package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
)

// AttachmentHandler serves transaction receipt (comprovante) uploads, listing,
// download and deletion. All endpoints are org-scoped via the tenant middleware.
type AttachmentHandler struct {
	attachments *services.AttachmentService
	maxBytes    int64
}

func NewAttachmentHandler(attachments *services.AttachmentService, maxAttachmentMB int) *AttachmentHandler {
	if maxAttachmentMB <= 0 {
		maxAttachmentMB = 10
	}
	return &AttachmentHandler{
		attachments: attachments,
		maxBytes:    int64(maxAttachmentMB) * 1024 * 1024,
	}
}

// Upload POST /transactions/{id}/attachments — multipart/form-data, field "file".
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	userID := middlewares.UserID(r.Context())
	txID, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}

	// Hard cap the request body so a viewer cannot exhaust memory/disk. A small
	// slack is added over maxBytes for the multipart envelope overhead.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes+(1<<20))
	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		response.Error(w, http.StatusRequestEntityTooLarge, "file_too_large",
			"Arquivo excede o tamanho máximo permitido")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request",
			"Arquivo ausente (campo multipart \"file\")")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = sniffContentType(file)
	}

	item, err := h.attachments.Upload(orgID, userID, txID, header.Filename, contentType, header.Size, file)
	if writeServiceError(w, err) {
		return
	}
	response.Created(w, item)
}

// List GET /transactions/{id}/attachments — any member.
func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	txID, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.attachments.List(orgID, txID)
	if writeServiceError(w, err) {
		return
	}
	response.OK(w, items)
}

// Download GET /attachments/{id}/download — streams the bytes with
// Content-Disposition. Any member (org-scoped).
func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	meta, rc, err := h.attachments.Download(orgID, id)
	if writeServiceError(w, err) {
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", meta.ContentType)
	if meta.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeHeaderFilename(meta.FileName)+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

// Delete DELETE /attachments/{id} — RequireWrite.
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middlewares.OrgID(r.Context())
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	if writeServiceError(w, h.attachments.Delete(orgID, id)) {
		return
	}
	response.NoContent(w)
}

// sniffContentType reads the first 512 bytes to detect the content type, then
// rewinds the reader so the upload reads the full content. multipart.File is a
// ReadSeeker, so the rewind is safe.
func sniffContentType(file io.ReadSeeker) string {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	ct := http.DetectContentType(buf[:n])
	_, _ = file.Seek(0, io.SeekStart)
	return ct
}

// sanitizeHeaderFilename strips characters that would break the quoted
// Content-Disposition filename (quotes and control-ish separators).
func sanitizeHeaderFilename(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch r {
		case '"', '\\', '\r', '\n':
			out = append(out, '_')
		default:
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return "arquivo"
	}
	return string(out)
}
