package validator

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = newValidator()

// newValidator builds the shared validator and teaches it to report errors under
// the field's JSON name. Without this, validator reports the Go struct field name
// ("ClosingDay"), which does not match the wire contract ("closing_day") and so
// cannot be matched to a form input by the SPA. Structs with no json tag keep the
// lower-cased Go field name, which is the historical behaviour.
func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return strings.ToLower(fld.Name)
		}
		return name
	})
	return v
}

// MaxJSONBody caps a JSON request body. Without a cap, a single field of a
// few hundred MB (even on the public /auth/login) was buffered whole in
// memory, and a handful of parallel requests ran the process out of memory.
const MaxJSONBody = 1 << 20

// BindJSON decodes the request body (up to MaxJSONBody) into dst and validates
// it using struct tags. It returns a map of field->message on failure (empty
// map = ok).
func BindJSON(r *http.Request, dst interface{}) (map[string]string, error) {
	return BindJSONLimit(r, dst, MaxJSONBody)
}

// BindJSONLimit is BindJSON with an explicit size cap, for the few routes that
// legitimately take more (e.g. committing a large statement import).
func BindJSONLimit(r *http.Request, dst interface{}, limit int64) (map[string]string, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, limit)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var grande *http.MaxBytesError
		if errors.As(err, &grande) {
			return map[string]string{"body": "Corpo da requisição muito grande"}, err
		}
		return map[string]string{"body": "JSON inválido"}, err
	}
	return Validate(dst), nil
}

// Validate runs struct validation and returns field errors keyed by the field's
// JSON name (see newValidator), so the SPA can attach each message to the input
// that produced it.
func Validate(dst interface{}) map[string]string {
	if err := v.Struct(dst); err != nil {
		fields := map[string]string{}
		for _, fe := range err.(validator.ValidationErrors) {
			fields[fe.Field()] = message(fe)
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
		return boundMessage("Mínimo", fe)
	case "max":
		return boundMessage("Máximo", fe)
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

// boundMessage phrases a min/max violation according to what the bound actually
// measures: the value itself for numbers, the length for strings, and the item
// count for collections. Saying "caracteres" for a numeric field (e.g.
// closing_day) misreports a value bound as a length bound.
func boundMessage(prefix string, fe validator.FieldError) string {
	switch fe.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return prefix + ": " + fe.Param()
	case reflect.Slice, reflect.Array, reflect.Map:
		return prefix + " de " + fe.Param() + " itens"
	default:
		return prefix + " de " + fe.Param() + " caracteres"
	}
}
