package repository

import (
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/deetoo/go-file-storage-service/models"
)

// FileData holds binary data and metadata of a file
type FileData struct {
	Data        []byte
	ContentType string
}

// FileMetadata stores the metadata of a file
type FileMetadata struct {
	ID          primitive.ObjectID `json:"_id"`
	ContentType string             `json:"_contentType"`
	MD5         string             `json:"md5"`
	Length      int                `json:"length"`
}

// FileRepository manages file persistence
type FileRepository interface {
	List(filter map[string]string, opts ...*ListOptions) ([]*models.File, error)
	GetByID(string, *models.File) error
	Create(*models.File, *FileData) error
	Update(*models.File, *FileData) error
	DeleteByID(string) error
	DownloadByID(string, io.Writer) (*FileMetadata, error)
}
