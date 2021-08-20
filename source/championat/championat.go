package championat

import (
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"log"
	"net/http"
	"strconv"
)

const (
	categoryCount  = 1
	layoutDateTime = "15:04:05 02.01.2006 -07:00"
	ru             = ""
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

	urlParts[0] = "https://championat.com/"
	for i := 0; i < categoryCount; i++ {
		urlParts[1] = cat[i].link
		lastLink := ""
		prevLastLink := ""
		for indexPage := 1; ; indexPage++ {
			// ====================================================================
			// make URL
			// ====================================================================
			newsUrl := urlParts[0] + urlParts[1] + "/" + strconv.Itoa(indexPage) + ".html"

			statusCode := http.StatusRequestTimeout
			for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
				statusCode, lastLink = NewsPageParser(repo, newsUrl, models.News{
					CatID:  cat[i].id,
					AuthID: newsInfo.AuthID,
					Image:  "",
				})
			}

			// =========================================================
			// if lastLink equal to prevLastLink then we got a end of news for this category
			// =========================================================
			if lastLink == prevLastLink {
				break
			}
			prevLastLink = lastLink

			if statusCode == http.StatusNotFound {
				break
			}
		}
	}
}

func getCategories(repo *repository.Repository) ([]Categories, error) {
	var category []Categories

	category = append(category, Categories{
		link: "news",
		name: "Спорт",
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