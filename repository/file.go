package repository

import (
	"github.com/deetoo/go-file-storage-service/models"
)

// FileData holds binary data and metadata of a file
type FileData struct {
	Data        []byte
	ContentType string
}

// FileRepository manages file persistence
type FileRepository interface {
	List(filter map[string]string, opts ...*ListOptions) ([]*models.File, error)
	GetByID(string, *models.File) error
	Create(*models.File, *FileData) error
	Update(*models.File, *FileData) error
	DeleteByID(string) error
}
