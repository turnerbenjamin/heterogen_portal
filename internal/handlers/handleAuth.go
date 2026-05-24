package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/go_gbf/internal/auth"
	"github.com/turnerbenjamin/go_gbf/internal/config"
	"github.com/turnerbenjamin/go_gbf/internal/db"
	"github.com/turnerbenjamin/go_gbf/internal/etc"
	"github.com/turnerbenjamin/go_gbf/internal/logging"
	"github.com/turnerbenjamin/go_gbf/internal/templates"
)

type JwtCookie struct {
	id        string
	createdOn time.Time
}

const (
	JWT_COOKIE_IDENTIFIER = "hg_login_jwt"
)

var (
	VALIDATION_MSG_GIVEN_NAME_EMPTY    = "Please provide a given name"
	VALIDATION_MSG_GIVEN_NAME_TOO_LONG = fmt.Sprintf("Given name cannot exceed %d characters", db.DB_CONSTRAINT_GIVEN_NAME_MAX)
	VALIDATION_MSG_LAST_NAME_EMPTY     = "Please provide a last name"
	VALIDATION_MSG_LAST_NAME_TOO_LONG  = fmt.Sprintf("Last name cannot exceed %d characters", db.DB_CONSTRAINT_FAMILY_NAME_MAX)
)

var (
	ErrServer = etc.ToastAndPageErrors(
		500,
		"An unexpected error has occurred. Please try again later",
		"An unexpected error has occurred. Please try again later",
	)
)

type UserRaft struct {
	User         *db.User
	ToastSuccess string
}

func POST_UserSignIn(appSettings config.AppSettings, ts *templates.Store, userRepo db.UserRepo) AppHandler {
	return func(w http.ResponseWriter, r *http.Request, l logging.Logger) etc.AppError {
		// Parse bearer token
		bearerToken := r.Header.Get("Authorization")
		tokenClaims, err := auth.ValidateToken(r.Context(), bearerToken)
		if err != nil {
			l.AddKV("Error", err.Error())
			return ErrServer
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
			l.AddKV("Failed to insert user", err.Error())
			return ErrServer
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
			return ErrServer
		}
		setJWTCookie(w, tokenString)

		w.Header().Add("HX-Push-Url", "/")
		conf := templates.PageConfig{
			ContentOnly:  r.Header.Get("HX-Request") != "",
			Title:        "HETEROGEN",
			ToastSuccess: "You've been signed-in successfully",
		}

		err = ts.Execute(
			templates.TMPL_PAGE_APP,
			w,
			templates.TemplateArgs{PageConfig: conf, Data: UserRaft{User: user}},
		)
		if err != nil {
			l.AddKV("server_error", err.Error())
			return ErrServer
		}
		return nil
	}
}

func NewParseJWTMiddleware(settings config.AppSettings, userRepo db.UserRepo) MiddlewareWithRaft[UserRaft] {
	return func(next AppHandlerWithRaft[UserRaft]) AppHandlerWithRaft[UserRaft] {
		return func(w http.ResponseWriter, r *http.Request, l logging.Logger, raft UserRaft) etc.AppError {
			if jwtCookie, err := r.Cookie(JWT_COOKIE_IDENTIFIER); err == nil {
				payload, ok := parseUserJwtCookie(jwtCookie, settings)
				if ok {
					l.AddKV("TOKEN ID", payload.id)
					user, err := userRepo.RetrieveUserById(payload.id)
					if err != nil {
						l.AddKV("User", "User not found")
						unsetJWTCookie(w)
					} else {
						l.AddKV("User", user.UserName)
						raft.User = user
					}
				}
			}
			return next(w, r, l, raft)
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
		Name:        JWT_COOKIE_IDENTIFIER,
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
		Name:        JWT_COOKIE_IDENTIFIER,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      -1,
		Expires:     time.Unix(0, 0),
		Secure:      true,
		Partitioned: true,
		HttpOnly:    true,
	})
}
