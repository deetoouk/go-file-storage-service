package controllers

import (
	"net/http"

	"github.com/deetoo/go-file-storage-service/models"
	"github.com/deetoo/go-file-storage-service/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileController implements the file routes.
type FileController struct {
	repo repository.FileRepository
}

// NewFileController returns a new file controller instance.
func NewFileController(repo repository.FileRepository) *FileController {
	return &FileController{
		repo: repo,
	}
}

// Get fetches a file by ID
func (i *FileController) Get(ctx *gin.Context) {
	file := &models.File{}

	err := i.repo.GetByID(ctx.Param("id"), file)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No file with such id"})
		return
	}

	ctx.JSON(http.StatusOK, file)
}

// Create adds a file
func (i *FileController) Create(ctx *gin.Context) {
	file := &models.File{}

	if err := ctx.BindJSON(file); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := i.repo.Create(file, &repository.FileData{})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, file)
}

// Update updates a file by ID
func (i *FileController) Update(ctx *gin.Context) {
	file := &models.File{}

	err := ctx.BindJSON(file)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file.ID, err = primitive.ObjectIDFromHex(ctx.Param("id"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = i.repo.Update(file, &repository.FileData{})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, &file)
}

// Delete removes a file by ID
func (i *FileController) Delete(ctx *gin.Context) {
	err := i.repo.DeleteByID(ctx.Param("id"))

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
