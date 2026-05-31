package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/finance-sh/finance-sh/internal/dto"
	"github.com/finance-sh/finance-sh/internal/middlewares"
	"github.com/finance-sh/finance-sh/internal/services"
	"github.com/finance-sh/finance-sh/pkg/response"
	"github.com/finance-sh/finance-sh/pkg/validator"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func reqMeta(r *http.Request) services.AuthMeta {
	return services.AuthMeta{UserAgent: r.UserAgent(), IP: r.RemoteAddr}
}

// Register POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	res, err := h.auth.Register(req, reqMeta(r))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRegistrationClosed):
			response.Error(w, http.StatusForbidden, "registration_closed", err.Error())
		case errors.Is(err, services.ErrEmailTaken):
			response.Error(w, http.StatusConflict, "email_taken", err.Error())
		case errors.Is(err, services.ErrUnsupportedCurrency):
			response.Error(w, http.StatusUnprocessableEntity, "unsupported_currency", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Não foi possível concluir o cadastro")
		}
		return
	}
	response.Created(w, res)
}

// RegistrationOpen GET /auth/registration-open — public flag so the frontend can
// hide the signup UI when self-service registration is disabled.
func (h *AuthHandler) RegistrationOpen(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]bool{"open": h.auth.RegistrationOpen()})
}

// Login POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	res, err := h.auth.Login(req, reqMeta(r))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		case errors.Is(err, services.ErrAccountDisabled):
			response.Error(w, http.StatusForbidden, "account_disabled", err.Error())
		case errors.Is(err, services.ErrAccountLocked):
			response.Error(w, http.StatusLocked, "account_locked",
				fmt.Sprintf("Conta temporariamente bloqueada por excesso de tentativas. Tente novamente em %d minutos.", h.auth.LockoutMinutes()))
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao autenticar")
		}
		return
	}

	// 2FA challenge: return only the mfa token, no session yet.
	if res.MFARequired {
		response.OK(w, dto.MFAChallengeResponse{MFARequired: true, MFAToken: res.MFAToken})
		return
	}
	response.OK(w, res.Auth)
}

// VerifyTwoFactor POST /auth/2fa/verify — exchanges an mfa_token + code for a
// full session.
func (h *AuthHandler) VerifyTwoFactor(w http.ResponseWriter, r *http.Request) {
	var req dto.TwoFactorVerifyRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	res, err := h.auth.VerifyTwoFactorLogin(req, reqMeta(r))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			response.Error(w, http.StatusUnauthorized, "invalid_token", "Token inválido ou expirado")
		case errors.Is(err, services.ErrInvalidCode):
			response.Error(w, http.StatusUnauthorized, "invalid_code", "Código de verificação inválido")
		default:
			response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao verificar o segundo fator")
		}
		return
	}
	response.OK(w, res)
}

// VerifyEmail POST /auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	if err := h.auth.VerifyEmail(req.Token); err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			response.Error(w, http.StatusBadRequest, "invalid_token", "Token inválido ou expirado")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao verificar o e-mail")
		return
	}
	response.OK(w, map[string]string{"message": "E-mail verificado com sucesso."})
}

// ResendVerification POST /auth/verify-email/resend — always returns 200 so the
// existence/verification status of an email is never revealed.
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendVerificationRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	h.auth.ResendVerification(req.Email)
	response.OK(w, map[string]string{
		"message": "Se o e-mail existir e não estiver verificado, enviaremos um novo link.",
	})
}

// Refresh POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	res, err := h.auth.Refresh(req.RefreshToken, reqMeta(r))
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid_token", "Token de atualização inválido ou expirado")
		return
	}
	response.OK(w, res)
}

// Logout POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}
	_ = h.auth.Logout(req.RefreshToken)
	response.NoContent(w)
}

// ForgotPassword POST /auth/forgot-password — always returns 200 with a generic
// message so the existence of an email is never revealed. In development the
// reset token is echoed back to ease testing.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	// The raw token is intentionally NOT returned to the client (no enumeration /
	// token leak). Without SMTP the operator gets the link from the server log or
	// the `-reset-password` CLI.
	if _, err := h.auth.ForgotPassword(req.Email); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao processar solicitação")
		return
	}

	// email_sent is instance-level (whether SMTP is configured), not per-account —
	// it does not reveal whether the e-mail exists, so it's safe to expose and
	// lets the UI show the right message (email vs. "ask the operator / see logs").
	response.OK(w, map[string]any{
		"message":    "Se o e-mail existir, enviaremos instruções de redefinição.",
		"email_sent": h.auth.MailEnabled(),
	})
}

// ResetPassword POST /auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if fields, err := validator.BindJSON(r, &req); err != nil || len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	if err := h.auth.ResetPassword(req.Token, req.Password); err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			response.Error(w, http.StatusBadRequest, "invalid_token", "Token inválido ou expirado")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "Erro ao redefinir a senha")
		return
	}
	response.OK(w, map[string]string{"message": "Senha redefinida com sucesso."})
}

// Me GET /me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.UserID(r.Context())
	res, err := h.auth.Me(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "not_found", "Usuário não encontrado")
		return
	}
	response.OK(w, res)
}
