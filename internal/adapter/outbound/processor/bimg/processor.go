package bimg

import (
	"bytes"
	"fmt"
	"io"

	"github.com/andreychano/compressor-golang/internal/core/domain"
	"github.com/h2non/bimg"
)

// Processor implements image compression using bimg/libvips.
type Processor struct{}

// NewProcessor creates a new bimg-based processor.
func NewProcessor() *Processor {
	return &Processor{}
}

// Supports reports whether the given MIME type is supported.
func (p *Processor) Supports(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

// Process compresses the input file according to the provided options.
func (p *Processor) Process(inputFile domain.File, opts domain.Options) (domain.File, error) {
	if _, err := inputFile.Content.Seek(0, 0); err != nil {
		return domain.File{}, fmt.Errorf("failed to seek file content: %w", err)
	}

	buffer, err := io.ReadAll(inputFile.Content)
	if err != nil {
		return domain.File{}, fmt.Errorf("cannot read: %w", err)
	}

	img := bimg.NewImage(buffer)

	processOptions := bimg.Options{
		Quality:       opts.Quality,
		StripMetadata: true,
	}

	processedBuffer, err := img.Process(processOptions)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to process image: %w", err)
	}

	return domain.File{
		Content:  bytes.NewReader(processedBuffer),
		MimeType: "image/" + opts.Format,
		Size:     int64(len(processedBuffer)),
	}, nil
}
