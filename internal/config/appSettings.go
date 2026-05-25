package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrInvalidDotenvFile              = errors.New("invalid dotenv file. Each line should contain a key and value separated by whitespace")
	ErrUnableToReadJWTSecret          = errors.New("unable to read JWT secret")
	ErrUnableToReadDBConnectionString = errors.New("unable to read DB connection string")
)

type AppSettings struct {
	JwtPrivateKey string
	SqlServerDsn  string
}

func GetAppSettings(ctx context.Context, dotenvPath string, isRunningLocally bool) (*AppSettings, error) {
	if isRunningLocally {
		err := parseDotEnv(dotenvPath)
		if err != nil {
			return nil, err
		}
	}

	settings := AppSettings{}

	var ok bool

	if settings.JwtPrivateKey, ok = os.LookupEnv("JWT_PRIVATE_KEY"); !ok {
		return nil, ErrUnableToReadJWTSecret
	}

	if settings.SqlServerDsn, ok = os.LookupEnv("SQL_SERVER_DSN"); !ok {
		return nil, ErrUnableToReadDBConnectionString
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
