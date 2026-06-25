package etc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrInvalidDotenvFile = errors.New("invalid dotenv file. Each line should contain a key and value separated by whitespace")

	ErrUnableToReadAppUrlBase             = errors.New("unable to read app url base")
	ErrUnableToReadAppJwtSecret           = errors.New("unable to read app jwt secret")
	ErrUnableToReadOidcStateSecret        = errors.New("unable to read oidc state secret")
	ErrUnableToReadDbConnectionString     = errors.New("unable to read db connection string")
	ErrUnableToReadUserPortalClientId     = errors.New("unable to read user portal client id")
	ErrUnableToReadUserPortalClientSecret = errors.New("unable to read user portal client secret")
	ErrUnableToReadUserPortalIssuerUrl    = errors.New("unable to read user portal issuer url")
)

type AppSettings struct {
	IsRunningLocally       bool
	AppUrlBase             string
	AppJwtSecret           []byte
	OidcStateSecret        []byte
	SqlServerDsn           string
	UserPortalClientId     string
	UserPortalClientSecret string
	UserPortalIssuerUrl    string
}

func GetAppSettings(
	ctx context.Context,
	dotenvPath string,
	isRunningLocally bool,
) (*AppSettings, error) {
	var ok bool
	settings := AppSettings{
		IsRunningLocally: isRunningLocally,
	}

	if settings.IsRunningLocally {
		err := parseDotEnv(dotenvPath)
		if err != nil {
			return nil, err
		}
	}

	if settings.AppUrlBase, ok = os.LookupEnv("APP_URL_BASE"); !ok {
		return nil, ErrUnableToReadAppUrlBase
	}

	appJWTSecret := ""
	if appJWTSecret, ok = os.LookupEnv("APP_JWT_SECRET"); !ok {
		return nil, ErrUnableToReadAppJwtSecret
	}
	settings.AppJwtSecret = []byte(appJWTSecret)

	oidcStateSecret := ""
	if oidcStateSecret, ok = os.LookupEnv("OIDC_STATE_SECRET"); !ok {
		return nil, ErrUnableToReadOidcStateSecret
	}
	settings.OidcStateSecret = []byte(oidcStateSecret)

	if settings.SqlServerDsn, ok = os.LookupEnv("SQL_SERVER_DSN"); !ok {
		return nil, ErrUnableToReadDbConnectionString
	}

	if settings.UserPortalClientId, ok = os.LookupEnv("USER_PORTAL_CLIENT_ID"); !ok {
		return nil, ErrUnableToReadUserPortalClientId
	}

	if settings.UserPortalClientSecret, ok = os.LookupEnv("USER_PORTAL_CLIENT_SECRET"); !ok {
		return nil, ErrUnableToReadUserPortalClientSecret
	}

	if settings.UserPortalIssuerUrl, ok = os.LookupEnv("USER_PORTAL_ISSUER_URL"); !ok {
		return nil, ErrUnableToReadUserPortalIssuerUrl
	}

	return &settings, nil
}

func parseDotEnv(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing dotenv file: %s", err.Error())
		}
	}()

	s := bufio.NewScanner(file)
	for s.Scan() {
		words := strings.SplitN(s.Text(), "=", 2)
		if len(words) != 2 {
			return ErrInvalidDotenvFile
		}

		key := strings.Trim(words[0], " ")
		value := strings.Trim(words[1], " ")

		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}

	return s.Err()
}
