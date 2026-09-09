package validator_test

import (
	"testing"

	"github.com/finance-sh/finance-sh/pkg/validator"
	"github.com/stretchr/testify/assert"
)

type sample struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Accept   bool   `validate:"eq=true"`
}

func TestValidateReportsFieldMessages(t *testing.T) {
	// All invalid: empty name, bad email, short password, terms not accepted.
	in := sample{
		Name:     "",
		Email:    "not-an-email",
		Password: "short",
		Accept:   false,
	}
	fields := validator.Validate(in)

	// Keys are the lower-cased struct field names.
	assert.Equal(t, "Campo obrigatório", fields["name"])
	assert.Equal(t, "E-mail inválido", fields["email"])
	assert.Equal(t, "Mínimo de 8 caracteres", fields["password"])
	assert.Equal(t, "É necessário aceitar os termos", fields["accept"])
}

func TestValidateRequiredVsFormat(t *testing.T) {
	// Email present but malformed -> email message (not required).
	in := sample{
		Name:     "Diego",
		Email:    "bad@",
		Password: "longenough",
		Accept:   true,
	}
	fields := validator.Validate(in)
	assert.Equal(t, "E-mail inválido", fields["email"])
	// Valid fields produce no entry.
	_, hasName := fields["name"]
	assert.False(t, hasName)
	_, hasPassword := fields["password"]
	assert.False(t, hasPassword)
	_, hasAccept := fields["accept"]
	assert.False(t, hasAccept)
}

type bounded struct {
	ClosingDay int      `validate:"required,min=1,max=31"`
	Nickname   string   `validate:"max=5"`
	TagIDs     []string `validate:"max=2"`
}

// A min/max bound on a number limits the value, not a character count, so the
// message must not talk about "caracteres" (the phrasing that made
// `closing_day: 32` report "Máximo de 31 caracteres").
func TestValidateBoundMessageMatchesFieldKind(t *testing.T) {
	fields := validator.Validate(bounded{
		ClosingDay: 32,
		Nickname:   "nome-grande-demais",
		TagIDs:     []string{"a", "b", "c"},
	})

	assert.Equal(t, "Máximo: 31", fields["closingday"])
	assert.Equal(t, "Máximo de 5 caracteres", fields["nickname"])
	assert.Equal(t, "Máximo de 2 itens", fields["tagids"])
}

// Day 31 is valid: internal/cards clamps it to the last day of short months.
func TestValidateAcceptsDay31(t *testing.T) {
	fields := validator.Validate(bounded{ClosingDay: 31})
	_, hasClosingDay := fields["closingday"]
	assert.False(t, hasClosingDay)
}

type tagged struct {
	AccountID  string `json:"account_id" validate:"required,uuid"`
	ClosingDay int    `json:"closing_day" validate:"required,min=1,max=31"`
	Notes      string `json:"notes,omitempty" validate:"max=5"`
	Internal   string `json:"-" validate:"required"`
	Untagged   string `validate:"required"`
}

// Field errors must be keyed by the JSON name, which is what the SPA registers
// its inputs under. Keying them by the Go field name ("closingday") left the
// messages unattachable to any input.
func TestValidateKeysErrorsByJSONName(t *testing.T) {
	fields := validator.Validate(tagged{ClosingDay: 32, Notes: "longo demais"})

	assert.Equal(t, "Campo obrigatório", fields["account_id"])
	assert.Equal(t, "Máximo: 31", fields["closing_day"])
	// The name stops at the first comma: "notes,omitempty" -> "notes".
	assert.Equal(t, "Máximo de 5 caracteres", fields["notes"])
	// json:"-" and untagged fields fall back to the lower-cased Go name.
	assert.Equal(t, "Campo obrigatório", fields["internal"])
	assert.Equal(t, "Campo obrigatório", fields["untagged"])

	// The old Go-field-name keys must be gone.
	_, hasOldKey := fields["closingday"]
	assert.False(t, hasOldKey)
	_, hasOldAccountKey := fields["accountid"]
	assert.False(t, hasOldAccountKey)
}

func TestValidateValidStructReturnsNil(t *testing.T) {
	in := sample{
		Name:     "Diego",
		Email:    "diego@finance.sh",
		Password: "supersecret",
		Accept:   true,
	}
	assert.Nil(t, validator.Validate(in))
}
