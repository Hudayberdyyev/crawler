package ixbt

import (
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"log"
	"net/http"
	"time"
)

const (
	categoryCount  = 2
	layoutDateTime = "15:04:05 02.01.2006 -07:00"
	ru             = ""
	layoutDatePage = "2006/01/02"
	layoutDate = "02.01.2006"
)

type Categories struct {
	link string
	name string
	id   int
}

var urlParts [2]string
var cat []Categories

func StartParser(repo *repository.Repository, newsInfo models.News) {
	cat, err := getCategories(repo)
	if err != nil {
		log.Printf("error with get category id: %v\n", err)
		return
	}

	urlParts[0] = "https://ixbt.com/"
	for i := 0; i < categoryCount; i++ {
		urlParts[1] = cat[i].link
		for indexPage := 1; ; indexPage++ {
			page := time.Now().AddDate(0, 0, -indexPage)
			dateStr := page.Format(layoutDatePage)
			// ====================================================================
			// make URL
			// ====================================================================
			newsUrl := urlParts[0] + urlParts[1] + "/" + dateStr + "/"

			newsId, err := repo.Database.GetLatestNewsIdByAuthorAndCategory(cat[i].id, newsInfo.AuthID)
			if err != nil {
				log.Printf("error with get latest news id: %v", err)
			}

			latestLink, err := repo.Database.GetLatestNewsUrlByNewsId(newsId)
			if err != nil {
				log.Printf("error with get latest new url: %v", err)
			}
			statusCode := NewsPageParser(repo, newsUrl, latestLink, models.News{
				CatID:  cat[i].id,
				AuthID: newsInfo.AuthID,
				Image:  "",
				PublishDate: page,
			})

			if statusCode == http.StatusNotModified || statusCode == http.StatusBadRequest {
				break
			}
		}
	}
}

func getCategories(repo *repository.Repository) ([]Categories, error) {
	var category []Categories

	category = append(category, Categories{
		link: "news",
		name: "Технология",
		id:   0,
	})

	category = append(category, Categories{
		link: "articles",
		name: "Публикации",
		id:   0,
	})

	for i := 0; i < categoryCount; i++ {
		id, err := repo.Database.GetCategoryIdByName(category[i].name)
		if err != nil {
			log.Printf("error with get category id by name=%s, error: %v\n", category[i].name, err)
			continue
		}
		category[i].id = id
	}

	return category, nil
}