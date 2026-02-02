package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andreychano/compressor-golang/internal/adapter/outbound/repository/local/pathvalidator"
	"github.com/andreychano/compressor-golang/internal/core/domain"
)

type LocalFileStorage struct {
	basePath      string
	pathValidator *pathvalidator.Validator
}

// NewLocalFileStorage initializes local storage and path validator.
func NewLocalFileStorage(basePath string) *LocalFileStorage {
	return &LocalFileStorage{
		basePath:      basePath,
		pathValidator: pathvalidator.New(basePath),
	}
}

func (s *LocalFileStorage) Save(ctx context.Context, file domain.File, relativePath string) (domain.SavedFile, error) {
	if err := s.pathValidator.Validate(relativePath); err != nil {
		return domain.SavedFile{}, fmt.Errorf("access denied: invalid path")
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return domain.SavedFile{}, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return domain.SavedFile{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err := file.Content.Seek(0, 0); err != nil {
		return domain.SavedFile{}, fmt.Errorf("failed to seek file content: %w", err)
	}

	if _, err := io.Copy(dst, file.Content); err != nil {
		return domain.SavedFile{}, fmt.Errorf("failed to write file content: %w", err)
	}

	info, err := dst.Stat()
	if err != nil {
		return domain.SavedFile{}, fmt.Errorf("failed to stat saved file: %w", err)
	}

	return domain.SavedFile{
		Path:           fullPath,
		CompressedSize: info.Size(), // file byte size
	}, nil
}

func (s *LocalFileStorage) Get(ctx context.Context, relativePath string) (domain.File, error) {
	if err := s.pathValidator.Validate(relativePath); err != nil {
		return domain.File{}, err
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	f, err := os.Open(fullPath)
	if err != nil {
		return domain.File{}, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return domain.File{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return domain.File{
		Content:  f,
		Size:     stat.Size(),
		MimeType: "application/octet-stream",
	}, nil
}
