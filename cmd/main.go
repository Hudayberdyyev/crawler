package main

import (
	"fmt"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"github.com/Hudayberdyyev/crawler/repository/postgres"
	"github.com/Hudayberdyyev/crawler/repository/storage"
	"github.com/Hudayberdyyev/crawler/source/TurkmenPortal"
	"github.com/Hudayberdyyev/crawler/source/championat"
	"github.com/Hudayberdyyev/crawler/source/ixbt"
	"github.com/Hudayberdyyev/crawler/source/rozetked"
	"github.com/Hudayberdyyev/crawler/source/wylsacom"
	"github.com/jackc/pgx"
	"github.com/minio/minio-go/v7"
	"log"
	"time"
)

const (
	ParsingInterval = 1 // on seconds
	TurkmenPortalID = 1
	Rozetked = 2
	Wylsacom = 3
	Championat = 4
	IXBT = 5
)

func main() {
	fmt.Println("crawler is starting ... ")

	db, err := initDB(postgres.Config{
		Host: "localhost",
		Username: "postgres",
		Password: "qwerty",
		Port: 5432,
		DBName: "postgres",
		SSLMode: "disable",
	})

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	minioClient, err := initMinio(storage.Config{
		Endpoint: "127.0.0.1:9000",
		AccessKeyId: "AHMET",
		SecretAccesKey: "Ah25101996!",
		UseSSL: false,
	})

	repository := repository.NewRepository(db, minioClient)

	if err != nil {
		log.Fatalf("Error with init minio: %v\n", err)
	}

	RunParser(repository, ParsingInterval)
}

func initDB(config postgres.Config)  (*pgx.Conn, error){

	db, err := postgres.NewPostgresDB(config)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func initMinio(config storage.Config) (*minio.Client, error) {
	return storage.NewMinio(config)
}

func RunParser(repo *repository.Repository, second int) {
	ticker := time.NewTicker(time.Duration(second) * time.Second)

	for _ = range ticker.C{
		// ============================================================
		fmt.Println("Checking updates [turkmenportal]")	
		TurkmenPortal.ParseTurkmenPortal(repo, models.News{
			CatID:  0,
			AuthID: TurkmenPortalID,
			Image:  "",
		})
		// ============================================================
		fmt.Println("Checking updates [rozetked]")
		rozetked.StartParser(repo, models.News{
			CatID:  0,
			AuthID: Rozetked,
			Image:  "",
		})
		// ============================================================
		fmt.Println("Checking updates [wylsacom]")
		wylsacom.StartParser(repo, models.News{
			CatID: 0,
			AuthID: Wylsacom,
			Image: "",
		})
		// ============================================================
		fmt.Println("Checking updates [championat]")
		championat.StartParser(repo, models.News{
			CatID:  0,
			AuthID: Championat,
			Image:  "",
		})
		// ============================================================
		fmt.Println("Checking updates [ixbt]")
		ixbt.StartParser(repo, models.News{
			CatID: 0,
			AuthID: IXBT,
			Image: "",
		})
		// ============================================================
		fmt.Println("everything up to date !!!")
	}
}