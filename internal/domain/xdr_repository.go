package domain

import (
	"context"
	"log"
	"log/slog"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// XDRRepository defines the interface for XDR operations
type XDRRepository interface {
	GetXDRList(ctx context.Context, iCustomer int, fromDateUnix, toDateUnix int64, page, pageSize int) (map[string]interface{}, error)
	GetXDRByIXDR(ctx context.Context, iXDR int) (bson.M, error)
	PostXDRList(ctx context.Context, data bson.M) (primitive.ObjectID, error)
	AcknowledgeXDRList(ctx context.Context, id primitive.ObjectID, s3Path string) error
}

// xdrRepository implements XDRRepository
type xdrRepository struct {
	collection *mongo.Collection
}

// NewXDRRepository creates a new XDRRepository
func NewXDRRepository(db *mongo.Database) XDRRepository {
	return &xdrRepository{
		collection: db.Collection("xdr_list"),
	}
}

// GetXDRList retrieves a paginated list of XDRs for a given customer.
func (repo *xdrRepository) GetXDRList(ctx context.Context, iCustomer int, fromDateUnix, toDateUnix int64, page, pageSize int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	query := bson.M{
		"i_customer":        iCustomer,
		"unix_connect_time": bson.M{"$gte": fromDateUnix, "$lte": toDateUnix},
	}

	// Count total documents
	total, err := repo.collection.CountDocuments(ctx, query)
	if err != nil {
		log.Printf("Error counting documents: %v", err)
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Find documents with pagination
	findOptions := options.Find().
		SetProjection(bson.M{"_id": 0}).
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.M{"unix_connect_time": -1}) // Sort by unix_connect_time descending

	cursor, err := repo.collection.Find(ctx, query, findOptions)
	if err != nil {
		log.Printf("Error fetching documents: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []bson.M
	if err = cursor.All(ctx, &records); err != nil {
		log.Printf("Error decoding documents: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"xdr_list":    records,
		"currentPage": page,
		"totalCount":  total,
		"totalPages":  totalPages,
	}

	return result, nil
}

// GetXDRByIXDR retrieves an XDR by its i_xdr value.
func (repo *xdrRepository) GetXDRByIXDR(ctx context.Context, iXDR int) (bson.M, error) {
	query := bson.M{"i_xdr": iXDR}
	var result bson.M

	err := repo.collection.FindOne(ctx, query, options.FindOne().SetProjection(bson.M{"_id": 0})).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error finding document: %v", err)
		return nil, err
	}

	return result, nil
}

// AcknowledgeXDRList updates the XDR record with the S3 path.
func (r *xdrRepository) AcknowledgeXDRList(ctx context.Context, id primitive.ObjectID, s3Path string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"s3_path": s3Path}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("Failed to update XDR record", "Error", err)
		return err
	}

	slog.Info("Successfully updated XDR record %v with S3 path: %s")
	return nil
}

// PostXDRList inserts an XDR record and returns its ID.
func (r *xdrRepository) PostXDRList(ctx context.Context, data bson.M) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		slog.Error("Failed to parse inserted ID")
		return primitive.NilObjectID, mongo.ErrNilDocument
	}

	return id, nil
}
