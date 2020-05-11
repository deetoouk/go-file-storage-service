package main

import (
	"github.com/deetoo/go-file-storage-service/controllers"
	"github.com/deetoo/go-file-storage-service/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

// GetRouter generates and manages all the application routes
func getRouter(client *mongo.Client, bucket *gridfs.Bucket) *gin.Engine {
	db := client.Database("files")

	fileRepository := repository.NewMongoFileRepository(db, bucket)

	fc := controllers.NewFileController(fileRepository)

	router := gin.Default()

	router.POST("/file", fc.Create)
	router.GET("/file/:id", fc.Get)
	router.PUT("/file/:id", fc.Update)
	router.DELETE("/file/:id", fc.Delete)

	return router
}
