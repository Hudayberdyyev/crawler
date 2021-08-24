package TurkmenPortal

import (
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"log"
	"net/http"
	"strconv"
)

const (
	tm = "tm"
	ru = ""
	categoryCount = 14
	layoutDateTime = "2006-01-02T15:04:05-07:00"
)

var urlParts [3]string
var Categories = [categoryCount]string{
	"politika-i-ekonomika",
	"kultura-i-obshchestvo",
	"Ykdysadyyetru",
	"obrazovanie",
	"turizm",
	"sport",
	"tehnologii",
	"avto",
	"biznes",
	"w-mire",
	"zdorove",
	"energetika",
	"Jemgiyetru",
	"compositions",
}


func ParseTurkmenPortal(repo *repository.Repository, newsInfo models.News) {
	for i := 0; i < categoryCount; i++ {
		if Categories[i] == "compositions" {
			urlParts[0] = "https://turkmenportal.com/"
			urlParts[2] = "/a/index?path=publikacii&Compositions_sort=date_added.desc&Blog_sort_temp=&page="
		} else {
			urlParts[0] = "https://turkmenportal.com/blog/a/index?path=novosti%2F"
			urlParts[2] = "&Blog_sort=date_added.desc&page="
		}
		// =========================================================
		// get latest success newsId by author and category (id)
		// =========================================================
		latestSuccessNewsId, err := repo.Database.GetLatestNewsIdByAuthorAndCategory(i+1, newsInfo.AuthID)
		if err != nil {
			latestSuccessNewsId, err = repo.Database.GetLatestNewsIdByAuthorAndCategory(i+1, newsInfo.AuthID)
			if err != nil {
				log.Printf("error with get lateset newsId by Author and CategoryId: %s\n", err.Error())
				return
			}
			err = repo.Database.SetLastUpdStatus(latestSuccessNewsId, 1)
			if err != nil {
				log.Printf("error with set lastUpd status to %d: %s\n", 1, err)
				return
			}
		}
		firstNewsId := 0
		for indexPage := 1; ; indexPage++ {
			urlParts[1] = Categories[i]
			newUrl := urlParts[0] + urlParts[1] + urlParts[2] + strconv.Itoa(indexPage)

			// =========================================================
			// we need cycle the parser until got a connection
			// =========================================================
			statusCode := http.StatusRequestTimeout
			for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
				statusCode = NewsPageParser(repo, newUrl, &firstNewsId, models.News{
					CatID:  i + 1,
					AuthID: newsInfo.AuthID,
					Image:  "",
				})
			}
			// =========================================================
			// if lastLink equal to prevLastLink then we got a end of news for this category
			// =========================================================

			if statusCode == http.StatusNotFound {
				break
			}

			if statusCode == http.StatusNotModified {
				err = repo.Database.SetLastUpdStatus(firstNewsId, 1)
				if err != nil {
					log.Printf("error with set lastUpdStatus for newsId(%d) : %s\n", firstNewsId, err.Error())
					return
				}
				break
			}
		}
	}
}