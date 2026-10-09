package middlewares

import (
	"net/http"
	"strings"

	"github.com/finance-sh/finance-sh/internal/repositories"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/google/uuid"
)

// AccountGate runs after Auth and enforces, on the server, two account states
// that only existed in the SPA or at login:
//
//   - a disabled account is refused at once, not when its 15-minute access
//     token happens to expire;
//   - an account that must change its password (temporary password from the
//     admin, or the seed) can only read /me and change the password. Before,
//     whoever held the temporary password used the whole API through it.
func AccountGate(users *repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserID(r.Context())
			if userID == uuid.Nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Não autenticado")
				return
			}
			user, err := users.FindByID(userID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Não autenticado")
				return
			}
			if user.Disabled {
				response.Error(w, http.StatusForbidden, "account_disabled", "Conta desativada. Contate o administrador.")
				return
			}
			if user.MustChangePassword && !liberadoSemTrocarSenha(r) {
				response.Error(w, http.StatusForbidden, "must_change_password", "Troque a senha antes de continuar.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func liberadoSemTrocarSenha(r *http.Request) bool {
	p := strings.TrimSuffix(r.URL.Path, "/")
	return (r.Method == http.MethodPost && p == "/api/v1/me/change-password") ||
		(r.Method == http.MethodGet && p == "/api/v1/me")
}
