package validator

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// BindJSON decodes the request body into dst and validates it using struct
// tags. It returns a map of field->message on failure (empty map = ok).
func BindJSON(r *http.Request, dst interface{}) (map[string]string, error) {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return map[string]string{"body": "JSON inválido"}, err
	}
	return Validate(dst), nil
}

// Validate runs struct validation and returns field errors.
func Validate(dst interface{}) map[string]string {
	if err := v.Struct(dst); err != nil {
		fields := map[string]string{}
		for _, fe := range err.(validator.ValidationErrors) {
			fields[strings.ToLower(fe.Field())] = message(fe)
		}
		return fields
	}
	return nil
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Campo obrigatório"
	case "email":
		return "E-mail inválido"
	case "min":
		return "Mínimo de " + fe.Param() + " caracteres"
	case "max":
		return "Máximo de " + fe.Param() + " caracteres"
	case "oneof":
		return "Valor inválido"
	case "eq":
		if fe.Param() == "true" {
			return "É necessário aceitar os termos"
		}
		return "Valor inválido"
	default:
		return "Valor inválido"
	}
}
