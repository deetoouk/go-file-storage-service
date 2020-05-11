package repository

import (
	"context"
	"crypto/md5"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/deetoo/go-file-storage-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoFileRepository hold the configuration for the mongo repo
type MongoFileRepository struct {
	collection *mongo.Collection
	bucket     *gridfs.Bucket
}

// NewMongoFileRepository create a new Mongo File Repository
func NewMongoFileRepository(db *mongo.Database, bucket *gridfs.Bucket) *MongoFileRepository {
	return &MongoFileRepository{
		collection: db.Collection("files"),
		bucket:     bucket,
	}
}

// List finds files by metadata
func (r *MongoFileRepository) List(metadata map[string]string, opts ...*ListOptions) ([]*models.File, error) {
	filter := map[string]string{}
	for k, v := range metadata {
		filter["metadata."+k] = v
	}

	ctx := context.Background()

	options := options.Find()

	options.SetLimit(100)
	options.SetSort(bson.D{{"created_at", -1}})

	for _, o := range opts {
		if o.Limit > 0 {
			options.SetLimit(o.Limit)

			if o.Page > 0 {
				options.SetSkip(o.Limit * o.Page)
			}
		}

		if o.OrderBy != "" {
			options.SetSort(bson.D{{o.OrderBy, o.OrderDirection}})
		}
	}

	cursor, err := r.collection.Find(ctx, filter, options)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	results := []*models.File{}

	for cursor.Next(ctx) {
		file := &models.File{}
		err := cursor.Decode(&file)
		if err != nil {
			return nil, err
		}

		results = append(results, file)
	}

	return results, nil
}

// GetByID gets a file fetched by it's id
func (r *MongoFileRepository) GetByID(id string, file *models.File) error {
	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	err = r.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(file)

	if err != nil {
		return fmt.Errorf("Document with ID: %v does not exist", id)
	}

	return nil
}

// Create creates a file and stores it in the database
func (r *MongoFileRepository) Create(file *models.File, fd *FileData) error {
	i, err := r.collection.InsertOne(context.Background(), file)

	if err != nil {
		return err
	}

	file.ID = i.InsertedID.(primitive.ObjectID)

	err = r.uploadFileAndAssociate(fd, file)

	if err != nil {
		r.DeleteByID(file.ID.Hex())
		return err
	}

	return r.GetByID(file.ID.Hex(), file)
}

// Update updates a file and stores it in the database
func (r *MongoFileRepository) Update(file *models.File, fd *FileData) error {
	if file.ID.IsZero() {
		return fmt.Errorf("File ID is required when updating a file")
	}

	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": file.ID}, bson.M{"$set": file})

	if err != nil {
		return err
	}

	err = r.GetByID(file.ID.Hex(), file)

	if err != nil {
		return err
	}

	if len(fd.Data) == 0 {
		return nil
	}

	err = r.uploadFileAndAssociate(fd, file)

	if err != nil {
		return err
	}

	return nil
}

// DeleteByID deletes a file by id
func (r *MongoFileRepository) DeleteByID(id string) error {
	file := &models.File{}

	err := r.GetByID(id, file)

	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(context.Background(), bson.M{"_id": file.ID})

	if err != nil {
		return err
	}

	return r.bucket.Delete(file.FileID)
}

func (r *MongoFileRepository) uploadFileAndAssociate(fd *FileData, file *models.File) error {
	opts := options.GridFSUpload().SetMetadata(bson.D{
		{"id", file.ID},
		{"contentType", fd.ContentType},
		{"md5", md5.Sum(fd.Data)},
		{"length", len(fd.Data)},
	})

	uploadStream, err := r.bucket.OpenUploadStream(
		file.Name,
		opts,
	)

	if err != nil {
		return err
	}

	defer uploadStream.Close()

	_, err = uploadStream.Write(fd.Data)

	if err != nil {
		return err
	}

	// Delete old file
	if !file.FileID.IsZero() {
		return r.bucket.Delete(file.FileID)
	}

	_, err = r.collection.UpdateOne(context.Background(), bson.M{"_id": file.ID}, bson.M{"$set": bson.M{"file_id": uploadStream.FileID}})

	if err != nil {
		return err
	}

	var ok bool

	file.FileID, ok = uploadStream.FileID.(primitive.ObjectID)

	if !ok {
		return fmt.Errorf("Could not convert FileID to ObjectID: %v", uploadStream.FileID)
	}

	return err
}
