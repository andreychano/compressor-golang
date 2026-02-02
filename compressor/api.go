package compressor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/andreychano/compressor-golang/internal/adapter/outbound/processor/bimg"
	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local"
	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

// Options describes compression settings exposed to library users.
type Options struct {
	Format    string // "jpeg", "png", "webp"
	Quality   int    // 1–100
	MaxWidth  int    // 0 = no limit (использовать дефолт из cfg.Image или без ограничения)
	MaxHeight int    // 0 = no limit
}

// Result contains metadata about the compressed image.
type Result struct {
	MimeType string
	Size     int64
}

// Compressor is a high-level façade for image compression.
type Compressor struct {
	svc *service.CompressionService
}

// New creates a new Compressor façade.
func New(svc *service.CompressionService) *Compressor {
	return &Compressor{svc: svc}
}

// Compress reads input, applies compression options and returns compressed data
// together with basic metadata.
func (c *Compressor) Compress(r io.Reader, opts Options) ([]byte, Result, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		return nil, Result{}, fmt.Errorf("failed to read input: %w", err)
	}

	// Определяем MIME по содержимому.
	mimeType := http.DetectContentType(buf.Bytes())

	file := domain.File{
		Content:  bytes.NewReader(buf.Bytes()),
		MimeType: mimeType,
		Size:     int64(buf.Len()),
	}

	domainOpts := domain.Options{
		Format:    opts.Format,
		Quality:   opts.Quality,
		MaxWidth:  opts.MaxWidth,
		MaxHeight: opts.MaxHeight,
	}

	outFile, err := c.svc.Process(file, domainOpts)
	if err != nil {
		return nil, Result{}, err
	}

	outBuf := &bytes.Buffer{}
	if _, err := io.Copy(outBuf, outFile.Content); err != nil {
		return nil, Result{}, fmt.Errorf("failed to read processed file: %w", err)
	}

	return outBuf.Bytes(), Result{
		MimeType: outFile.MimeType,
		Size:     outFile.Size,
	}, nil
}

// NewDefault creates a Compressor with default bimg processor
// and local filesystem storage under the given base path.
func NewDefault(basePath string) *Compressor {
	proc := bimg.NewProcessor()
	repo := local.NewLocalFileStorage(basePath)

	cfg := config.Config{
		Storage: config.Storage{
			Path:             basePath,
			CompressedSubdir: "compressed",
			TmpSubdir:        "tmp",
		},
		Image: config.Image{
			DefaultFormat:  "jpeg",
			DefaultQuality: 80,
			MaxWidth:       3840,
			MaxHeight:      2160,
			AllowFormats:   []string{"jpeg", "png", "webp"},
		},
	}

	svc := service.NewCompressionService(repo, cfg, proc)

	return &Compressor{svc: svc}
}
