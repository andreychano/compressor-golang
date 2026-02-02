package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/andreychano/compressor-golang/internal/config"
	"github.com/andreychano/compressor-golang/internal/core/domain"
	portmocks "github.com/andreychano/compressor-golang/internal/core/port/mocks"
	"github.com/andreychano/compressor-golang/internal/core/service"
)

func TestCompressionService_CompressAndSave_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := portmocks.NewMockFileRepository(ctrl)
	processorMock := portmocks.NewMockProcessor(ctrl)

	// минимальный конфиг для теста
	cfg := config.Config{
		Storage: config.Storage{
			CompressedSubdir: "compressed",
		},
		Image: config.Image{
			DefaultFormat:  "jpeg",
			DefaultQuality: 50,
			MaxWidth:       3840,
			MaxHeight:      2160,
		},
	}

	s := service.NewCompressionService(repoMock, cfg, processorMock)

	file := domain.File{MimeType: "image/png"}
	opts := domain.Options{Format: "jpeg"}

	processorMock.EXPECT().
		Supports(file.MimeType).
		Return(true)

	compressed := domain.File{MimeType: "image/jpeg"}

	processorMock.EXPECT().
		Process(file, gomock.Any()).
		Return(compressed, nil)

	repoMock.EXPECT().
		Save(gomock.Any(), compressed, gomock.Any()).
		Return(domain.SavedFile{
			Path:           "compressed/some-id.jpeg",
			CompressedSize: 123,
		}, nil)

	saved, err := s.CompressAndSave(context.Background(), file, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.Path == "" {
		t.Fatalf("expected non-empty path")
	}
}
