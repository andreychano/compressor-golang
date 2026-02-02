package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/port"
	"github.com/google/uuid"
)

type CompressionService struct {
	processors []port.Processor
	repository port.FileRepository
	cfg        config.Config
}

func NewCompressionService(repo port.FileRepository, cfg config.Config, processors ...port.Processor) *CompressionService {
	return &CompressionService{
		repository: repo,
		processors: processors,
		cfg:        cfg,
	}
}

func (s *CompressionService) Process(file domain.File, opts domain.Options) (domain.File, error) {
	var selectedProcessor port.Processor

	for _, p := range s.processors {
		if p.Supports(file.MimeType) {
			selectedProcessor = p
			break
		}
	}

	if selectedProcessor == nil {
		return domain.File{}, fmt.Errorf("unsupport file type: %s", file.MimeType)
	}

	return selectedProcessor.Process(file, opts)
}

func (s *CompressionService) CompressAndSave(
	ctx context.Context,
	file domain.File,
	reqOpts domain.Options,
) (domain.SavedFile, error) {
	opts := domain.Options{
		Format:    s.cfg.Image.DefaultFormat,
		Quality:   s.cfg.Image.DefaultQuality,
		MaxWidth:  s.cfg.Image.MaxWidth,
		MaxHeight: s.cfg.Image.MaxHeight,
	}

	if reqOpts.Format != "" {
		opts.Format = reqOpts.Format
	}
	if reqOpts.Quality != 0 {
		opts.Quality = reqOpts.Quality
	}
	if reqOpts.MaxWidth != 0 {
		opts.MaxWidth = reqOpts.MaxWidth
	}
	if reqOpts.MaxHeight != 0 {
		opts.MaxHeight = reqOpts.MaxHeight
	}

	compressedFile, err := s.Process(file, opts)
	if err != nil {
		return domain.SavedFile{}, err
	}

	uniqueID := uuid.New().String()
	fileName := fmt.Sprintf("%s.%s", uniqueID, opts.Format)
	filePath := filepath.Join(s.cfg.Storage.CompressedSubdir, fileName)

	return s.repository.Save(ctx, compressedFile, filePath)
}

func (s *CompressionService) GetFile(ctx context.Context, path string) (domain.File, error) {
	return s.repository.Get(ctx, path)
}
