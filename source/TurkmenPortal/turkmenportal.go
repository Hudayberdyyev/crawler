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
		for indexPage := 1; ; indexPage++ {
			urlParts[1] = Categories[i]
			newUrl := urlParts[0] + urlParts[1] + urlParts[2] + strconv.Itoa(indexPage)

			newsId, err := repo.GetLatestNewsIdByAuthorAndCategory(i+1, newsInfo.AuthID)
			if err != nil {
				log.Printf("error with get latest news id: %v", err)
			}

			latestLink, err := repo.GetLatestNewsUrlByNewsId(newsId)
			if err != nil {
				log.Printf("error with get latest new url: %v", err)
			}

			statusCode := NewsPageParser(repo, newUrl, latestLink, models.News{
				CatID:  i + 1,
				AuthID: newsInfo.AuthID,
				Image:  "",
			})

			if statusCode == http.StatusNotModified || statusCode == http.StatusBadRequest{
				break
			}
		}
	}
}