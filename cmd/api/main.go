package main

import (
	"context"
	"net/http"

	httpadapter "github.com/andreychano/compressor-golang/internal/adapter/inbound/http"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/service"
	applogger "github.com/andreychano/compressor-golang/internal/logger"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad(ctx)

	applogger.Init(cfg.Log.ToGotoolsConfig())

	// --- ОТЛАДКА (ВРЕМЕННО) ---
	// Это покажет нам, что реально загрузилось

	storage := local.NewLocalFileStorage(cfg.Storage.Path)
	processor := bimg.NewProcessor()
	svc := service.NewCompressionService(storage, *cfg, processor)

	mux := http.NewServeMux()
	h := httpadapter.NewHandler(svc)
	h.RegisterRoutes(mux)

	maxBytes := cfg.HTTP.MaxUploadSizeBytes()
	handler := httpadapter.MaxUploadSize(maxBytes, mux)

	applogger.Log.Info().
		Str("address", cfg.HTTP.Address).
		Int64("max_bytes", maxBytes).
		Str("udp", cfg.Log.UDPAddress).
		Msg("Starting server")

	if err := http.ListenAndServe(cfg.HTTP.Address, handler); err != nil {
		applogger.Log.Error().
			Err(err).
			Msg("failed to start server")
	}
}
