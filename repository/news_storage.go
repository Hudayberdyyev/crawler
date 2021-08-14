package repository

import (
	"context"
	"github.com/minio/minio-go/v7"
	"io"
	"log"
	"net/http"
)

type NewsStorage struct {
	minioClient *minio.Client
}

func NewNewsStorage(client *minio.Client) *NewsStorage {
	return &NewsStorage{ minioClient: client }
}

func (n *NewsStorage) UploadImage(ctx context.Context,bucketName string, filePath string, objectName string) error {
	location := "us-east-1"

	err := n.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location, ObjectLocking: false})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := n.minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			// log.Printf("We already own %s\n", bucketName)
		} else {
			return err
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	imageReader, err := getImageReader(filePath)
	if err != nil {
		return err
	}

	_, err = n.minioClient.PutObject(ctx, bucketName, objectName, imageReader, -1, minio.PutObjectOptions{ContentType:"application/octet-stream"})
	if err != nil {
		return err
	}
	// fmt.Printf("Successfully uploaded bytes: %s\n", filePath)

	return nil
}

func (n *NewsStorage) RemoveImage(ctx context.Context,bucketName string, objectName string) error {
	err := n.minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{GovernanceBypass: true})
	if err != nil {
		return err
	}
	// fmt.Printf("Successfully removed object: %s\n", objectName)
	return nil
}

func getImageReader(URL string) (io.Reader, error) {
	if resp, err := http.Get(URL); err != nil {
		return nil, err
	} else {
		return resp.Body, nil
	}
}
