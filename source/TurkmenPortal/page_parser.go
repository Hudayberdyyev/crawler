package TurkmenPortal

import (
	"fmt"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
	"time"
)

func NewsPageParser(repo *repository.Repository, URL string, newsInfo models.News ) (int) {
	// ====================================================================
	// http get URL
	// ====================================================================
	fmt.Println(URL)
	res, err := http.Get(URL)

	if err != nil {
		log.Printf("http.Get(URL) error: %v\n", err)
		if strings.Contains(err.Error(), "no such host"){
			return http.StatusRequestTimeout
		}
		return http.StatusGatewayTimeout
	}

	if res.StatusCode == http.StatusNotFound {
		return http.StatusNotFound
	}

	defer res.Body.Close()

	// ====================================================================
	// Load the HTML document
	// ====================================================================
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Error load html document: %v\n", err)
		return http.StatusInternalServerError
	}

	// news list parse
	// ====================================================================

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
		imageLink, ok := s.Find("img").Eq(0).Attr("src")
		if !ok {
			log.Printf("No news images\n")
			continue
		}

		if strings.Contains(imageLink, "http") {
			newsInfo.Image = imageLink
		} else {
			newsInfo.Image="https://turkmenportal.com" + imageLink
		}

		// link article
		// ====================================================================
		link, ok := s.Find("h4.entry-title > a").Eq(0).Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}

		// publishDate article
		// ====================================================================
		publishDateStr, ok := s.Find("time").Eq(0).Attr("datetime")
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
		// ===================================================================
		_, err = repo.Database.GetNewsIdByUrl(link)
		if err == nil {
			log.Printf("%s link already has in database\n", link)
			return http.StatusNotModified
		}

		// parsing ru version of news
		// ====================================================================
		statusCode := http.StatusRequestTimeout
		for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
			statusCode = NewsContentParser(repo, newsInfo, models.NewsText{
				NewsID: 0,
				Hl:     ru,
				Title:  "",
				Url:    link,
			})
		}
		// parsing tm version of news
		// ====================================================================
		statusCode  = http.StatusRequestTimeout
		for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
			statusCode = NewsContentParser(repo, newsInfo, models.NewsText{
				NewsID: 0,
				Hl:     tm,
				Title:  "",
				Url:    link,
			})
		}
	}

	return http.StatusOK
}
