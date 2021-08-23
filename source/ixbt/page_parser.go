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

func NewsPageParser(repo *repository.Repository, URL string, newsInfo models.News) (int, string) {
	// ====================================================================
	// http get URL
	// ====================================================================
	fmt.Println(URL)
	res, err := http.Get(URL)

	if err != nil {
		log.Printf("http.Get(URL) error: %v\n", err)
		if strings.Contains(err.Error(), "no such host"){
			return http.StatusRequestTimeout, "error"
		}
		return http.StatusGatewayTimeout, "error"
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return http.StatusNotFound, "error"
	}

	// ====================================================================
	// Load the HTML document
	// ====================================================================
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Error load html document: %v\n", err)
		return http.StatusInternalServerError, "error"
	}
	// ====================================================================
	// news list parse
	// ====================================================================
	var lastLink string

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
		link, ok := s.Find("a.item__text--title").Eq(0).Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}
		link = "https://ixbt.com" + link
		lastLink = link

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
		if publishDate.Year() < 2020 {
			return http.StatusNotFound, "very_old"
		}

		// ====================================================================
		// checking a new article
		// ====================================================================
		_, err = repo.Database.GetNewsIdByUrl(link)
		if err == nil {
			log.Printf("%s link already has in database\n", link)
			continue
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

		statusCode := http.StatusRequestTimeout
		for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
			statusCode = NewsContentParser(repo, models.NewsText{
				NewsID: newsId,
				Hl:     ru,
				Title:  "",
				Url:    link,
			})
		}
	}

	return http.StatusOK, lastLink
}