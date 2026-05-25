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

	"github.com/turnerbenjamin/go_gbf/internal/app"
	"github.com/turnerbenjamin/go_gbf/internal/config"
	"github.com/turnerbenjamin/go_gbf/internal/db"
	"github.com/turnerbenjamin/go_gbf/internal/templates"
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
	appConfig, err := config.GetAppSettings(ctx, dotenvPath, isRunningLocally)

	if err != nil {
		return err
	}

	ts, err := templates.MakeTemplateStore(embeddedFiles, "templates", config.TemplateDataMap)
	if err != nil {
		return err
	}

	db_conn, err := db.SetUpDB(appConfig.SqlServerDsn)
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

	adminRepo := db.BuildUserRepo(db_conn, crypt{})
	defer adminRepo.Close()

	srv, err := app.NewServer(*appConfig, ts, embeddedFiles, adminRepo)
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
