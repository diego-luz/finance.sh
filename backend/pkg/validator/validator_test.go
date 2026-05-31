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

func TestValidateValidStructReturnsNil(t *testing.T) {
	in := sample{
		Name:     "Diego",
		Email:    "diego@finance.sh",
		Password: "supersecret",
		Accept:   true,
	}
	assert.Nil(t, validator.Validate(in))
}
