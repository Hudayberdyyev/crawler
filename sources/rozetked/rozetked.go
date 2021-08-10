package rozetked

import (
	"context"
	"fmt"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	categoryCount = 2
	layoutDateTime = "15:04:05 02.01.2006 -07:00"
)

type Categories struct {
	link string
	name string
	id 	 int
}

var urlParts [2]string
var cat []Categories

func NewsPageParser(repo *repository.Repository, URL string, latestLink string, newsInfo models.News ) int {
	// ====================================================================
	// http get URL
	// ====================================================================
	res, err := http.Get(URL)

	if err != nil {
		log.Printf("http.Get(URL) error: %v\n", err)
		return http.StatusBadRequest
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Status code error: %d %s\n", res.StatusCode, res.Status)
		return http.StatusBadRequest
	}

	// ====================================================================
	// Load the HTML document
	// ====================================================================
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Error load html document: %v\n", err)
		return http.StatusBadRequest
	}
	// ====================================================================
	// news list parse
	// ====================================================================
	var result []string
	var ids	   []int

	fmt.Println(URL)
	sel := doc.Find("div#app > div.container.r > div.r_content > div.home > div.home_left > div.home_left__posts").Children().Filter("div.post_new")
	for i := range sel.Nodes {
		s := sel.Eq(i)
		// ====================================================================
		// article parse
		// ====================================================================

		// ====================================================================
		// image
		// ====================================================================
		imageLink, ok := s.Find("img").Attr("src")
		if !ok {
			log.Printf("No news images\n")
			continue
		}
		newsInfo.Image="https://rozetked.me" + imageLink

		// ====================================================================
		// link
		// ====================================================================
		link, ok := s.Find("div.post_new-title > a").Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}

		// ====================================================================
		// publishDate
		// ====================================================================
		publishDateStr := s.Find("div.post_new-meta > div.post_new-meta-author > span").Text()

		splitDate := strings.Split(publishDateStr, " ")

		if cap(splitDate) == 2 {
			splitPublishTime := splitDate[0]
			splitPublishDate := splitDate[1]
			splitPublishTime = splitPublishTime + ":00"
			publishDateStr = splitPublishTime + " " + splitPublishDate + " +03:00"
		} else {
			splitPublishTime := splitDate[0]
			timeType := splitDate[1]

			var durationType string
			if strings.Contains(timeType, "час") { durationType = "h" } else {
				if strings.Contains(timeType, "мин") { durationType = "m" } else {
					if strings.Contains(timeType, "сек") { durationType = "s" }
				}
			}
			duration1, err := time.ParseDuration("-" + splitPublishTime + durationType)
			if err != nil {
				log.Printf("error with parse duration %s: %v\n", splitPublishTime, err)
			}
			publishDateStr = time.Now().Add(duration1).Format(layoutDateTime)
		}

		publishDate, err := time.Parse(layoutDateTime, publishDateStr)
		if err != nil {
			log.Printf("error with parse date: %v\n", err)
		}
		newsInfo.PublishDate = publishDate

		// ====================================================================
		// checking a new article
		// ====================================================================
		if link == latestLink {
			fmt.Println("everything up to date !")
			return http.StatusNotModified
		}

		// ====================================================================
		//	article to db
		// ====================================================================
		newsId, e := repo.CreateNews(newsInfo)
		if e != nil {
			log.Printf("Error with create news: %v\n", e)
			continue
		}

		// ====================================================================
		// image article to storage
		// ====================================================================
		uploadErr := repo.UploadImage(context.Background(), "news", newsInfo.Image, strconv.Itoa(newsId))
		if uploadErr != nil {
			log.Printf("error with upload image: %v\n", uploadErr)
		}

		// ====================================================================
		// add ids and links articles to slices
		// ====================================================================
		ids = append(ids, newsId)
		result = append(result, link)
	}

	// ====================================================================
	// iterate articles (tm, ru)
	// ====================================================================
	//for index, link := range result {
	//	NewsContentParser(repo, models.NewsText{
	//		NewsID: ids[index],
	//		Hl:     ru,
	//		Title:  "",
	//		Url:    link,
	//	})
	//}

	return http.StatusOK
}

func StartParser(repo *repository.Repository, newsInfo models.News) {
	cat, err := getCategories(repo)
	if err != nil {
		log.Printf("error with get category id: %v\n", err)
		return
	}

	urlParts[0] = "https://rozetked.me/"
	for i := 0; i < categoryCount; i++ {
		urlParts[1] = cat[i].link
		for indexPage := 1; ; indexPage++ {
			// ====================================================================
			// make URL
			// ====================================================================
			newsUrl := urlParts[0] + urlParts[1] + "?page=" + strconv.Itoa(indexPage)

			newsId, err := repo.Database.GetLatestNewsIdByAuthorAndCategory(i+1, newsInfo.AuthID)
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
			})

			if statusCode == http.StatusNotModified || statusCode == http.StatusBadRequest{
				break
			}
		}
	}
}

func getCategories(repo *repository.Repository) ([]Categories, error){
	var category []Categories

	category = append(category, Categories{
		link: "news",
		name: "Технология",
		id: 0,
	})

	category = append(category, Categories{
		link: "article",
		name: "Публикации",
		id: 0,
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