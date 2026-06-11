package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/auth"
	"github.com/turnerbenjamin/heterogen_portal/internal/config"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
)

type JwtCookie struct {
	id        string
	createdOn time.Time
}

const (
	jwtCookieIdentifier = "hg_login_jwt"
)

var (
	VALIDATION_MSG_GIVEN_NAME_EMPTY    = "Please provide a given name"
	VALIDATION_MSG_GIVEN_NAME_TOO_LONG = fmt.Sprintf("Given name cannot exceed %d characters", db.DB_CONSTRAINT_GIVEN_NAME_MAX)
	VALIDATION_MSG_LAST_NAME_EMPTY     = "Please provide a last name"
	VALIDATION_MSG_LAST_NAME_TOO_LONG  = fmt.Sprintf("Last name cannot exceed %d characters", db.DB_CONSTRAINT_FAMILY_NAME_MAX)
)

type UserState struct {
	User         *db.User
	ToastSuccess string
}

func POST_UserSignIn(appSettings config.AppSettings, ts *templates.Store, userRepo db.UserRepo) AppHandler[NoState] {
	return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[NoState]) *etc.AppError {
		// Parse bearer token
		bearerToken := r.Header.Get("Authorization")
		tokenClaims, err := auth.ValidateToken(r.Context(), bearerToken)
		if err != nil {
			return etc.NewServerError(err)
		}

		user, err := userRepo.UpsertUser(
			r.Context(),
			tokenClaims.Oid,
			tokenClaims.GivenName,
			tokenClaims.FamilyName,
			tokenClaims.UserName,
			tokenClaims.EmailAddress,
		)
		if err != nil {
			return etc.NewServerError(err)
		}

		// Create new JWT Token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		})

		tokenString, err := token.SignedString([]byte(appSettings.JwtPrivateKey))
		if err != nil {
			return etc.NewServerError(err)
		}
		setJWTCookie(w, tokenString)

		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return nil
	}
}

func NewParseJWTMiddleware(settings config.AppSettings, userRepo db.UserRepo) Middleware[UserState] {
	return func(next AppHandler[UserState]) AppHandler[UserState] {
		return func(w http.ResponseWriter, r *http.Request, c *PipelineContext[UserState]) *etc.AppError {
			if jwtCookie, err := r.Cookie(jwtCookieIdentifier); err == nil {
				payload, ok := parseUserJwtCookie(jwtCookie, settings)
				if ok {
					c.AddLoggerKV(slog.String("TOKEN ID", payload.id))

					user, err := userRepo.RetrieveUserById(payload.id)
					if err != nil {
						c.AddLoggerKV(slog.String("User", "User not found"))
						unsetJWTCookie(w)
					} else {
						c.AddLoggerKV(slog.String("User", user.UserName))
						c.state.User = user
					}
				}
			}
			return next(w, r, c)
		}
	}
}

func parseUserJwtCookie(jwtCookie *http.Cookie, settings config.AppSettings) (*JwtCookie, bool) {
	token, err := jwt.ParseWithClaims(
		jwtCookie.Value,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(settings.JwtPrivateKey), nil
		})
	if err != nil {
		return nil, false
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return nil, false
	}

	createdOn, err := token.Claims.GetIssuedAt()
	if err != nil {
		return nil, false
	}

	return &JwtCookie{
		id:        id,
		createdOn: createdOn.Time,
	}, true
}

func setJWTCookie(w http.ResponseWriter, tokenString string) {
	http.SetCookie(w, &http.Cookie{
		Name:        jwtCookieIdentifier,
		Value:       tokenString,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      int((time.Hour * 24) / time.Second),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}

func unsetJWTCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:        jwtCookieIdentifier,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      -1,
		Expires:     time.Unix(0, 0).UTC(),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}
