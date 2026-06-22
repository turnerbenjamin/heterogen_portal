package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turnerbenjamin/heterogen_portal/internal/app"
	"github.com/turnerbenjamin/heterogen_portal/internal/db"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
	"github.com/turnerbenjamin/heterogen_portal/internal/services"
	"github.com/turnerbenjamin/heterogen_portal/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates/* static/*
var embeddedFiles embed.FS

type crypt struct{}

func (c crypt) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

func (c crypt) CompareHashAndPassword(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

type tokenSignerAndParser struct {
}

func (sp *tokenSignerAndParser) Sign(token *jwt.Token, key []byte) (string, error) {
	return token.SignedString(key)
}

func (sp *tokenSignerAndParser) ParseWithClaims(
	tokenString string,
	claims jwt.Claims,
	keyFunc jwt.Keyfunc,
	parserOptions ...jwt.ParserOption,
) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		tokenString,
		claims,
		keyFunc,
		parserOptions...,
	)
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	var isRunningLocally bool
	flag.BoolVar(&isRunningLocally, "local", false, "specify that running locally")
	flag.Parse()

	dotenvPath, err := filepath.Abs("cmd/.env")
	if err != nil {
		return err
	}
	privateKeyPath, err := filepath.Abs("private-key.pem")
	if err != nil {
		return err
	}

	publicCertPath, err := filepath.Abs("public-cert.pem")
	if err != nil {
		return err
	}

	appSettings, err := etc.GetAppSettings(
		ctx,
		dotenvPath,
		privateKeyPath,
		publicCertPath,
		isRunningLocally,
	)

	if err != nil {
		return err
	}

	ts, err := templates.MakeTemplateStore(embeddedFiles, "templates", templates.TemplateDataMap)
	if err != nil {
		return err
	}

	db_conn, err := db.SetUpDB(appSettings.SqlServerDsn)
	if err != nil {
		return err
	}
	defer func() {
		fmt.Println("closing DB connection")
		err = db_conn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error closing DB connection: %s\n", err)
		}
	}()

	userRepo := db.BuildUserRepo(db_conn, crypt{})
	defer userRepo.Close()

	jwtSignerAndParser := &tokenSignerAndParser{}
	authService, err := services.NewAuthService(
		ctx,
		appSettings,
		&http.Client{},
		userRepo,
		jwtSignerAndParser,
	)
	if err != nil {
		return err
	}

	srv, err := app.NewServer(
		appSettings,
		ts,
		embeddedFiles,
		authService,
	)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", "8080"),
		Handler: srv,
	}
	go func() {
		fmt.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		fmt.Println("\nshutting down http server")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})

	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
