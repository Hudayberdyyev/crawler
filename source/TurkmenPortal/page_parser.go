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

func NewsPageParser(repo *repository.Repository, URL string, firstNewsId *int, newsInfo models.News ) (int) {
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

	// ====================================================================
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

		// ===================================================================
		// checking a new article
		// ===================================================================
		checkId, err := repo.Database.GetNewsIdByUrl(link)
		if err == nil {
			// ===================================================================
			// if we don't set firstNewsId then setting firstNewsIf
			// for know at where the next time we started
			// ===================================================================
			if *firstNewsId == 0 { *firstNewsId = checkId }
			// log.Printf("%s link already has in database\n", link)
			// ===================================================================
			// get lastUpdStatus for this news
			// ===================================================================
			lastUpdStatus, err := repo.Database.GetLastUpdStatus(checkId)
			if err != nil {
				log.Printf("error with get lastUpdStatus for newsId(%d): %s\n", checkId, err.Error())
			}
			// ===================================================================
			// if status equal 1 then set statusValue to 0, and return this func with NotModified status
			// ===================================================================
			if lastUpdStatus == 1 {
				err = repo.Database.SetLastUpdStatus(checkId, 0)
				if err != nil {
					log.Printf("error with set lastUpdStatus for newsId(%d): %s\n", checkId, err.Error())
				}
				return http.StatusNotModified
			}
			continue
		}

		// parsing ru version of news
		// ====================================================================
		statusCode := http.StatusRequestTimeout
		for statusCode == http.StatusRequestTimeout || statusCode == http.StatusGatewayTimeout {
			statusCode = NewsContentParser(repo, newsInfo, firstNewsId, models.NewsText{
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
			statusCode = NewsContentParser(repo, newsInfo, firstNewsId, models.NewsText{
				NewsID: 0,
				Hl:     tm,
				Title:  "",
				Url:    link,
			})
		}
	}

	return http.StatusOK
}
