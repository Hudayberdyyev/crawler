package repository

import (
	"context"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/jackc/pgx"
	"github.com/minio/minio-go/v7"
)

type Database interface {
	GetLatestNewsIdByAuthorAndCategory (catId, authId int) (int, error)
	GetLatestNewsUrlByNewsId(newsId int) (string, error)
	CreateNews(newsInfo models.News) (int, error)
	CreateNewsText(newsText models.NewsText) (int, error)
	CreateNewsContent(newsContent models.NewsContent) (int, error)
}

type Storage interface {
	UploadImage(ctx context.Context,bucketName string, filePath string, objectName string) error
}

type Repository struct {
	Database
	Storage
}

func NewRepository(db *pgx.Conn, minioClient *minio.Client) *Repository {
	return &Repository{
		Database:	NewNewsDatabase(db),
		Storage: 	NewNewsStorage(minioClient),
	}
}