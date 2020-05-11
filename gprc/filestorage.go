package grpc

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/deetoo/go-file-storage-service/filestorage"
	"github.com/deetoo/go-file-storage-service/models"
	"github.com/deetoo/go-file-storage-service/repository"
)

var (
	mB          = 1 << 20 // 1Mb
	maxFileSize = 128 * mB
)

// RequestFileData contains the Raw file data
type RequestFileData struct {
	Data        *bytes.Buffer
	ContentType string
	Size        int
}

// Write populates the struct with the raw data
func (rfd *RequestFileData) Write(chunk []byte, contentType string) error {
	if contentType != "" {
		rfd.ContentType = contentType
	}

	length := len(chunk)

	if length > 0 {
		rfd.Size += length

		if rfd.Size > maxFileSize {
			return status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", rfd.Size, maxFileSize)
		}
		_, err := rfd.Data.Write(chunk)

		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}
	}

	return nil
}

// FileStorageServer hold the configuration for the file storage
type FileStorageServer struct {
	repo repository.FileRepository
}

// NewFileStorageServer creates a new instance of the NewFileStorage server
func NewFileStorageServer(repo repository.FileRepository) *FileStorageServer {
	return &FileStorageServer{repo: repo}
}

// Find searches for a list of files and returns them
func (s *FileStorageServer) Find(req *filestorage.FindRequest, res filestorage.FileStorage_FindServer) error {
	files, err := s.repo.List(req.GetMetadata(), &repository.ListOptions{Limit: 100})

	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	for _, f := range files {
		res.Send(convertFile(f))
	}

	return nil
}

// Get fetches a single file
func (s *FileStorageServer) Get(ctx context.Context, r *filestorage.GetRequest) (*filestorage.File, error) {
	id := r.GetId()
	file := &models.File{}
	err := s.repo.GetByID(id, file)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "File with ID %v does not exist", id)
	}

	return convertFile(file), nil
}

// Upload stores a file using stream of chunks
func (s *FileStorageServer) Upload(stream filestorage.FileStorage_UploadServer) error {

	var file *models.File

	requestFileData := &RequestFileData{
		Data:        &bytes.Buffer{},
		ContentType: "",
		Size:        0,
	}

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			if file == nil {
				return status.Error(codes.InvalidArgument, "File was never passed")
			}

			if requestFileData.Data.Len() == 0 {
				return status.Error(codes.InvalidArgument, "File content never passed")
			}

			if requestFileData.ContentType == "" {
				return status.Error(codes.InvalidArgument, "File content type never passed")
			}

			s.repo.Create(file, &repository.FileData{
				Data:        requestFileData.Data.Bytes(),
				ContentType: requestFileData.ContentType,
			})

			stream.SendAndClose(convertFile(file))
			break
		}

		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
			return status.Error(codes.Internal, "Error reading stream")
		}

		if data := req.GetData(); data != nil {
			err := requestFileData.Write(data.GetChunk(), data.GetContentType())

			if err != nil {
				return err
			}
		}

		if reqFile := req.GetFile(); reqFile != nil {
			file = &models.File{
				Name:        reqFile.Name,
				Description: reqFile.Description,
				Metadata:    reqFile.Metadata,
			}
		}
	}

	return nil
}

// Replace stores and replaces a file using stream of chunks
func (s *FileStorageServer) Replace(stream filestorage.FileStorage_ReplaceServer) error {
	var file *models.File

	requestFileData := &RequestFileData{
		Data:        &bytes.Buffer{},
		ContentType: "",
		Size:        0,
	}

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			if file == nil {
				return status.Error(codes.InvalidArgument, "File was never passed")
			}

			if requestFileData.Data.Len() > 0 && requestFileData.ContentType == "" {
				return status.Error(codes.InvalidArgument, "File content type never passed")
			}

			s.repo.Update(file, &repository.FileData{
				Data:        requestFileData.Data.Bytes(),
				ContentType: requestFileData.ContentType,
			})

			stream.SendAndClose(convertFile(file))

			break
		}

		if err != nil {
			log.Fatalf("Error while reading client stream %v", err)
			return status.Error(codes.Internal, "Error reading stream")
		}

		if data := req.GetData(); data != nil {
			err := requestFileData.Write(data.GetChunk(), data.GetContentType())

			if err != nil {
				return err
			}
		}

		if reqFile := req.GetFile(); reqFile != nil {
			file = &models.File{
				Name:        reqFile.Name,
				Description: reqFile.Description,
				Metadata:    reqFile.Metadata,
			}

			file.ID, err = primitive.ObjectIDFromHex(reqFile.Id)

			if err != nil {
				return status.Errorf(codes.Internal, "Could not convert ID to Object ID: %v", err)
			}
		}
	}

	return nil
}

// Delete deletes a file
func (s *FileStorageServer) Delete(ctx context.Context, r *filestorage.DeleteRequest) (*filestorage.DeleteResponse, error) {
	err := s.repo.DeleteByID(r.GetId())

	if err != nil {
		return nil, err
	}

	return &filestorage.DeleteResponse{}, nil
}

// Download downloads a file
func (s *FileStorageServer) Download(r *filestorage.DownloadRequest, stream filestorage.FileStorage_DownloadServer) error {
	buf := &bytes.Buffer{}

	metadata, err := s.repo.DownloadByID(r.GetId(), buf)

	if err != nil {
		return status.Error(codes.NotFound, err.Error())
	}

	stream.Send(&filestorage.DownloadResponse{
		Data: &filestorage.FileData{
			Data: &filestorage.FileData_ContentType{
				ContentType: metadata.ContentType,
			},
		},
	})

	chunk := make([]byte, 1<<16)

	for {
		n, err := buf.Read(chunk)

		if err == io.EOF {
			break
		}

		stream.Send(&filestorage.DownloadResponse{
			Data: &filestorage.FileData{
				Data: &filestorage.FileData_Chunk{
					Chunk: chunk[:n],
				},
			},
		})
	}

	return nil
}

func toFile(file *filestorage.File) *models.File {
	result := &models.File{}

	result.Name = file.Name
	result.Description = file.Description
	result.Metadata = file.Metadata

	return result
}

func convertFile(file *models.File) *filestorage.File {
	target := &filestorage.File{}

	target.Id = file.ID.Hex()
	target.Name = file.Name
	target.Description = file.Description
	target.Metadata = file.Metadata
	target.UpdatedAt = file.UpdatedAt.Format(time.RFC3339)
	target.CreatedAt = file.CreatedAt.Format(time.RFC3339)

	return target
}
