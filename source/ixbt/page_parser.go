package ixbt

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

func NewsPageParser(repo *repository.Repository, URL string, latestLink string, newsInfo models.News) int {
	// ====================================================================
	// http get URL
	// ====================================================================
	fmt.Println(URL)
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
	var ids []int

	sel := doc.Find("div.b-block.block__newslistdefault.b-lined-title > ul").Children().Filter("li")
	for i := range sel.Nodes {
		s := sel.Eq(i)
		// ====================================================================
		// article parse
		// ====================================================================

		// ====================================================================
		// image default
		// ====================================================================
		newsInfo.Image = "https://www.ixbt.com/images/ixbt-logo-new-sm.jpg"

		// ====================================================================
		// link
		// ====================================================================
		link, ok := s.Find("a.item__text--title").Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}
		link = "https://ixbt.com" + link

		// ====================================================================
		// publishDate
		// ====================================================================
		metaSel := s.Find("span").Eq(0)
		publishTimeStr := strings.Trim(metaSel.Text(), " \n\t\r")

		publishTimeStr += ":00"
		publishDateStr := newsInfo.PublishDate.Format(layoutDate)
		publishDateStr = publishTimeStr + " " + publishDateStr + " +03:00"
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
		newsId, e := repo.Database.CreateNews(newsInfo)
		if e != nil {
			log.Printf("Error with create news: %v\n", e)
			continue
		}

		// ====================================================================
		// image article to storage
		// ====================================================================
		uploadErr := repo.Storage.UploadImage(context.Background(), "news", newsInfo.Image, strconv.Itoa(newsId))
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
	for index, link := range result {
		NewsContentParser(repo, models.NewsText{
			NewsID: ids[index],
			Hl:     ru,
			Title:  "",
			Url:    link,
		})
	}
	return http.StatusOK
}