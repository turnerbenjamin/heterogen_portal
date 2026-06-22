package etc

import (
	"bufio"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrUnableToCastRsaPrivateKey = errors.New("unable to cast RSA private key")

	ErrInvalidDotenvFile = errors.New("invalid dotenv file. Each line should contain a key and value separated by whitespace")

	ErrUnableToReadAppUrlBase                 = errors.New("unable to read app url base")
	ErrUnableToReadAppJWTSecret               = errors.New("unable to read app jwt secret")
	ErrUnableToReadOIDCStateSecret            = errors.New("unable to read oidc state secret")
	ErrUnableToReadDBConnectionString         = errors.New("unable to read db connection string")
	ErrUnableToReadUserPortalClientId         = errors.New("unable to read user portal client id")
	ErrUnableToReadUserPortalOAuthUrl         = errors.New("unable to read user portal oauth url")
	ErrUnableToReadUserPortalGetOIDCConfigUrl = errors.New("unable to read user portal get oidc config url")
)

type AppSettings struct {
	IsRunningLocally           bool
	AppUrlBase                 string
	AppJWTSecret               []byte
	OIDCStateSecret            []byte
	SqlServerDsn               string
	UserPortalClientId         string
	UserPortalOAuthUrl         string
	UserPortalGetOidcConfigUrl string
	RsaKey                     *rsa.PrivateKey
	Cert                       *x509.Certificate
}

func GetAppSettings(ctx context.Context, dotenvPath string, privateKeyPath string, publicCertPath string, isRunningLocally bool) (*AppSettings, error) {
	var ok bool
	settings := AppSettings{
		IsRunningLocally: isRunningLocally,
	}

	// Read private key for authenticating with Azure
	pemBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if settings.RsaKey, ok = privateKey.(*rsa.PrivateKey); !ok {
		return nil, ErrUnableToReadAppJWTSecret
	}

	// Read public certificate for authenticating with Azure
	certBytes, err := os.ReadFile(publicCertPath)
	if err != nil {
		return nil, err
	}

	block, _ = pem.Decode(certBytes)
	if block == nil {
		return nil, errors.New("invalid cert PEM")
	}

	settings.Cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	// Load environment Variables
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
		return nil, ErrUnableToReadAppJWTSecret
	}
	settings.AppJWTSecret = []byte(appJWTSecret)

	oidcStateSecret := ""
	if appJWTSecret, ok = os.LookupEnv("OIDC_STATE_SECRET"); !ok {
		return nil, ErrUnableToReadOIDCStateSecret
	}
	settings.OIDCStateSecret = []byte(oidcStateSecret)

	if settings.SqlServerDsn, ok = os.LookupEnv("SQL_SERVER_DSN"); !ok {
		return nil, ErrUnableToReadDBConnectionString
	}

	if settings.UserPortalClientId, ok = os.LookupEnv("USER_PORTAL_CLIENT_ID"); !ok {
		return nil, ErrUnableToReadUserPortalClientId
	}

	if settings.UserPortalOAuthUrl, ok = os.LookupEnv("USER_PORTAL_OAUTH_URL"); !ok {
		return nil, ErrUnableToReadUserPortalOAuthUrl
	}

	if settings.UserPortalGetOidcConfigUrl, ok = os.LookupEnv("USER_PORTAL_GET_OIDC_CONFIG_URL"); !ok {
		return nil, ErrUnableToReadUserPortalGetOIDCConfigUrl
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
