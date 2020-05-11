package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/deetoo/go-file-storage-service/filestorage"
	grpcServer "github.com/deetoo/go-file-storage-service/gprc"
	"github.com/deetoo/go-file-storage-service/repository"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
)

func main() {
	// Populates database with dummy data
	err := godotenv.Load()

	wg := sync.WaitGroup{}

	if err != nil {
		log.Fatal("Could not load dotenv file")
	}

	client := dbConn()
	defer client.Disconnect(context.TODO())

	bucket, _ := gridfs.NewBucket(
		client.Database("files"),
	)

	fileRepository := repository.NewMongoFileRepository(client.Database("files"), bucket)

	wg.Add(2)

	go func() {
		port, err := strconv.Atoi(os.Getenv("GRPC_PORT"))

		if err != nil {
			log.Fatal("failed to decode GRPC port from ENV var")
		}

		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		var opts []grpc.ServerOption

		server := grpc.NewServer(opts...)

		fileStorageServer := grpcServer.NewFileStorageServer(fileRepository)

		filestorage.RegisterFileStorageServer(server, fileStorageServer)

		fmt.Println("GRPC Server up and running on http://0.0.0.0:" + os.Getenv("GRPC_PORT"))

		reflection.Register(server)

		server.Serve(lis)

		wg.Done()
	}()

	go func() {
		fmt.Println("Server up and running on http://0.0.0.0:" + os.Getenv("PORT"))
		getRouter(client, bucket).Run()
		wg.Done()
	}()

	wg.Wait()
}
