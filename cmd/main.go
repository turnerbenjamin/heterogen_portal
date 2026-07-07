package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/app"
	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
)

//go:embed templates static
var embeddedFiles embed.FS

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	dotenvPath, err := filepath.Abs("cmd/.env")
	if err != nil {
		return err
	}

	appSettings, err := etc.GetAppSettings(
		ctx,
		dotenvPath,
	)
	if err != nil {
		return err
	}

	application, err := app.Init(ctx, appSettings, embeddedFiles)
	if err != nil {
		return err
	}
	defer application.CleanUp()

	go func() {
		fmt.Printf("listening on %s\n", application.Address())
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
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
		if err := application.Shutdown(shutdownCtx); err != nil {
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
