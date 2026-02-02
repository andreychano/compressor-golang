package domain

import "io"

type File struct {
	Content  io.ReadSeeker // Re-readable file content stream
	MimeType string        // MIME type, e.g. "image/jpeg"
	Size     int64         // File size in bytes
}

// Options defines parameters for compression operations.
type Options struct {
	Format    string // Target format (e.g. "webp", "jpeg")
	Quality   int    // Compression quality
	MaxWidth  int    // Maximum width in pixels
	MaxHeight int    // Maximum height in pixels
}

// SaveResult describes result of compress+save operation.
type SavedFile struct {
	Path           string // Full path where file is stored (e.g. storage/compressed/...)
	CompressedSize int64  // Size of compressed file in bytes
}
