package repo

import (
	"errors"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"gorm.io/gorm"
)

type Gormable interface {
	TableName() string
}

func CreateOne[T Gormable](app *global.App, doc T) (*T, error) {
	if err := app.Validator.Struct(doc); err != nil {
		return nil, err
	}
	db := app.DB

	if err := db.Create(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func GetOne[T Gormable](app *global.App, filter any) (*T, error) {
	db := app.DB
	if filter != nil {
		db = db.Where(filter)
	}
	var result T
	if err := db.First(&result).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("not found")
	} else if err != nil {
		return nil, err
	}
	return &result, nil
}

type Options struct {
	Sort   []string
	Omit   []string
	Limit  *uint
	Offset *uint
	Filter map[string]any
}

func GetMany[T Gormable](app *global.App, opts *Options) ([]T, error) {
	if opts == nil {
		opts = &Options{}
		opts.Filter = map[string]any{}
	}
	db := app.DB
	if len(opts.Filter) > 0 {
		db = db.Where(opts.Filter)
	}
	if len(opts.Omit) > 0 {
		db = db.Omit(opts.Omit...)
	}
	if len(opts.Sort) > 0 {
		db = db.Order(opts.Sort[0])
	}
	if opts.Limit != nil {
		db = db.Limit(int(*opts.Limit))
	}
	if opts.Offset != nil {
		db = db.Offset(int(*opts.Offset))
	}
	var results []T
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func Count[T Gormable](app *global.App, filter any) (int64, error) {
	db := app.DB.Model(new(T))
	if filter != nil {
		db = db.Where(filter)
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateOne[T Gormable](app *global.App, filter any, updates any) (*T, error) {
	db := app.DB

	var result T
	if filter != nil {
		db = db.Where(filter)
	}

	// Execute the update
	if err := db.Model(&result).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Re-fetch the updated record
	if filter != nil {
		if err := db.Where(filter).First(&result).Error; err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func InsertMany[T Gormable](app *global.App, docs []T) ([]T, error) {
	db := app.DB
	validator := app.Validator
	for i := range docs {
		if err := validator.Struct(&docs[i]); err != nil {
			return nil, err
		}
	}
	if err := db.Create(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func DeleteOne[T Gormable](app *global.App, filter any) error {
	db := app.DB
	if filter != nil {
		db = db.Where(filter)
	}
	result := db.Delete(new(T))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}
