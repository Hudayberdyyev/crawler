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
	GetCategoryIdByName(categoryName string) (int, error)
	GetTagIdByName(tagName string) (int, error)
	CreateTags(tagName string, hl string) (int, error)
	CreateTagsText(tagId int, tagName string, hl string) (int, error)
	CreateNewsTags(newsId int, tagId int) (int, error)
	UpdateNewsImageById(newsId int, imageLink string) (error)
	UpdateTagByContentId(contentId int, tagName string) (error)
}

type Storage interface {
	UploadImage(ctx context.Context,bucketName string, filePath string, objectName string) error
	RemoveImage(ctx context.Context,bucketName string, objectName string) error
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