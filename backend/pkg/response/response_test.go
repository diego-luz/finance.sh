package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	return env
}

func TestOK(t *testing.T) {
	rec := httptest.NewRecorder()
	response.OK(rec, map[string]string{"hello": "world"})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	env := decode(t, rec)
	assert.True(t, env.Success)
	assert.Nil(t, env.Error)
	data, ok := env.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "world", data["hello"])
}

func TestCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	response.Created(rec, map[string]int{"id": 7})

	assert.Equal(t, http.StatusCreated, rec.Code)
	env := decode(t, rec)
	assert.True(t, env.Success)
}

func TestPaginated(t *testing.T) {
	rec := httptest.NewRecorder()
	response.Paginated(rec, []string{"a", "b"}, map[string]int{"total": 2})

	assert.Equal(t, http.StatusOK, rec.Code)
	env := decode(t, rec)
	assert.True(t, env.Success)
	assert.NotNil(t, env.Data)
	assert.NotNil(t, env.Meta)
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	response.NoContent(rec)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	response.Error(rec, http.StatusNotFound, "not_found", "Recurso não encontrado")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	env := decode(t, rec)
	assert.False(t, env.Success)
	assert.Nil(t, env.Data)
	require.NotNil(t, env.Error)
	assert.Equal(t, "not_found", env.Error.Code)
	assert.Equal(t, "Recurso não encontrado", env.Error.Message)
	assert.Empty(t, env.Error.Fields)
}

func TestValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	fields := map[string]string{"email": "E-mail inválido"}
	response.ValidationError(rec, fields)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	env := decode(t, rec)
	assert.False(t, env.Success)
	require.NotNil(t, env.Error)
	assert.Equal(t, "validation_error", env.Error.Code)
	assert.Equal(t, "Dados inválidos", env.Error.Message)
	assert.Equal(t, "E-mail inválido", env.Error.Fields["email"])
}
