package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/covrom/geoip/internal/addr"
	"github.com/covrom/geoip/internal/handler"
)

func main() {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(l)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go addr.RegularUpdate(ctx, 12*time.Hour, wg)

	listen := ":8000"
	if s := os.Getenv("LISTEN_ADDR"); s != "" {
		listen = s
	}

	srv := http.Server{
		Addr:           listen,
		Handler:        handler.New(),
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("served", "listen", listen)
		slog.Error("server stopped", "err", srv.ListenAndServe())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	wg.Wait()
}
