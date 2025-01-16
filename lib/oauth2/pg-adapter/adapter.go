package pgadapter

import (
	"context"

	"gorm.io/gorm"
)

type Provider interface {
	Exec(ctx context.Context, query string, args ...interface{}) error
	SelectOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
}

func NewInstance(DB *gorm.DB) Provider {
	return &impl{
		db: DB,
	}
}

type impl struct{
	db *gorm.DB
}


func (i impl) Exec(ctx context.Context, query string, args ...interface{}) error {
	return i.db.Exec(query, args...).Error
}

func (i impl) SelectOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return i.db.Raw(query, args...).Scan(&dst).Error
}
