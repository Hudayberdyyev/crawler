package storage

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint string
	AccessKeyId string
	SecretAccesKey string
	UseSSL bool
}

func NewMinio(cfg Config) (*minio.Client, error){
	return minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyId, cfg.SecretAccesKey, ""),
		Secure: cfg.UseSSL,
	})
}