package repo

import (
	"context"
	"errors"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongoable interface {
	CollectionName() string
}

func collection[T Mongoable](app *global.App, doc T) *mongo.Collection {
	return app.Mongo.Database(app.MongoDBName).Collection(doc.CollectionName())
}

func MongoCreateOne[T Mongoable](app *global.App, doc T) (*T, error) {
	if err := app.Validator.Struct(doc); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection(app, doc).InsertOne(ctx, doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func MongoGetOne[T Mongoable](app *global.App, filter bson.M) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc T
	if err := collection(app, doc).FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &doc, nil
}

type MongoOptions struct {
	Sort   bson.D
	Limit  *int64
	Offset *int64
	Filter bson.M
}

func MongoGetMany[T Mongoable](app *global.App, opts *MongoOptions) ([]T, error) {
	if opts == nil {
		opts = &MongoOptions{}
	}
	if opts.Filter == nil {
		opts.Filter = bson.M{}
	}

	findOpts := options.Find()
	if len(opts.Sort) > 0 {
		findOpts.SetSort(opts.Sort)
	}
	if opts.Limit != nil {
		findOpts.SetLimit(*opts.Limit)
	}
	if opts.Offset != nil {
		findOpts.SetSkip(*opts.Offset)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc T
	cursor, err := collection(app, doc).Find(ctx, opts.Filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func MongoUpdateOne[T Mongoable](app *global.App, filter bson.M, updates bson.M) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc T
	col := collection(app, doc)
	if err := col.FindOneAndUpdate(ctx, filter, bson.M{"$set": updates}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &doc, nil
}

func MongoDeleteOne[T Mongoable](app *global.App, filter bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc T
	result, err := collection(app, doc).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("not found")
	}
	return nil
}

func MongoInsertMany[T Mongoable](app *global.App, docs []T) ([]T, error) {
	validator := app.Validator
	for i := range docs {
		if err := validator.Struct(&docs[i]); err != nil {
			return nil, err
		}
	}

	if len(docs) == 0 {
		return docs, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	anyDocs := make([]any, len(docs))
	for i := range docs {
		anyDocs[i] = docs[i]
	}

	if _, err := collection(app, docs[0]).InsertMany(ctx, anyDocs); err != nil {
		return nil, err
	}
	return docs, nil
}
