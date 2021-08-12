package TurkmenPortal

import (
	"context"
	"fmt"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"time"
)

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

	// news list parse
	// ====================================================================
	var result []string
	var ids	   []int

	sel := doc.Find("div#yw3 > div.items").Children()
	for i := range sel.Nodes {
		s := sel.Eq(i)
		if s.Nodes[0].Data != "article" {
			continue
		}

		// article parse
		// ====================================================================

		// image article
		// ====================================================================
		imageLink, ok := s.Find("img").Attr("src")
		if !ok {
			log.Printf("No news images\n")
			continue
		}
		newsInfo.Image="https://turkmenportal.com" + imageLink

		// link article
		// ====================================================================
		link, ok := s.Find("h4.entry-title > a").Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}

		// publishDate article
		// ====================================================================
		publishDateStr, ok := s.Find("time").Attr("datetime")
		if !ok {
			log.Printf("No date in news\n")
			continue
		}

		publishDate, err := time.Parse(layoutDateTime, publishDateStr)
		if err != nil {
			log.Printf("error with parse date: %v\n", err)
		}
		newsInfo.PublishDate = publishDate

		// checking a new article
		// ====================================================================
		if link == latestLink {
			fmt.Println("everything up to date !")
			return http.StatusNotModified
		}

		//	article to db
		// ====================================================================
		newsId, e := repo.CreateNews(newsInfo)
		if e != nil {
			log.Printf("Error with create news: %v\n", e)
			continue
		}

		// image article to storage
		// ====================================================================
		uploadErr := repo.UploadImage(context.Background(), "news", newsInfo.Image, strconv.Itoa(newsId))
		if uploadErr != nil {
			log.Printf("error with upload image: %v\n", uploadErr)
		}

		// add ids and links articles to slices
		// ====================================================================
		ids = append(ids, newsId)
		result = append(result, link)
	}
	// iterate articles (tm, ru)
	// ====================================================================
	for index, link := range result {
		NewsContentParser(repo, models.NewsText{
			NewsID: ids[index],
			Hl:     ru,
			Title:  "",
			Url:    link,
		})
		NewsContentParser(repo, models.NewsText{
			NewsID: ids[index],
			Hl:     tm,
			Title:  "",
			Url:    link,
		})
	}

	return http.StatusOK
}