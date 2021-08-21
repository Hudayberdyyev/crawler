package championat

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
	fmt.Println(URL)
	sel := doc.Find("div.page-content > div.page-main > div.news._all > div.news-items").Children()
	var postDate, lastLink string
	for i := range sel.Nodes {
		s := sel.Eq(i)
		// ====================================================================
		// if tag is date then get date
		// ====================================================================

		if s.Nodes[0].Attr[0].Val == "news-items__head" {
			splitPostDate := strings.Split(strings.Trim(s.Text(), " \n\t\r"), " ")
			if cap(splitPostDate) != 3 {
				log.Println("error date format: ", splitPostDate)
				continue
			}
			postDay, postMonth, postYear := splitPostDate[0], getMonthByRussianName(strings.ToLower(splitPostDate[1])), splitPostDate[2]
			if postMonth == "impossible" {
				log.Printf("impossible month name: %s\n", strings.ToLower(splitPostDate[1]))
				continue
			}
			if len(postDay) < 2 {
				postDay = "0" + postDay
			}
			if len(postMonth) < 2 {
				postMonth = "0" + postMonth
			}
			postDate = postDay + "." + postMonth + "." + postYear
			continue
		}

		// ====================================================================
		// if tag is article then get info
		// ====================================================================
		if s.Nodes[0].Attr[0].Val == "news-item" {

			// ====================================================================
			// image
			// ====================================================================
			newsInfo.Image = "https://championat.com/static/i/svg/logo.svg"

			// ====================================================================
			// link
			// ====================================================================
			link, ok := s.Find("div.news-item__content > a").Eq(0).Attr("href")
			if !ok {
				log.Printf("No news link\n")
				continue
			}
			if strings.Contains(link, "http") || link[0] != 47{
				continue
			}
			link = "https://championat.com" + link
			lastLink = link

			// ====================================================================
			// publishDate
			// ====================================================================
			metaSel := s.Find("div.news-item__time")
			publishTimeStr := strings.Trim(metaSel.Text(), " \n\t\r") + ":00"
			splitTime := strings.Split(publishTimeStr, ":")
			if cap(splitTime) > 1 {
				if len(splitTime[0]) < 2 {
					splitTime[0] = "0" + splitTime[0]
				}
				if len(splitTime[1]) < 2 {
					splitTime[1] = "0" + splitTime[1]
				}
				publishTimeStr = strings.Join(splitTime, ":")
			}
			publishDateStr := publishTimeStr + " " + postDate + " +03:00"

			publishDate, err := time.Parse(layoutDateTime, publishDateStr)
			if err != nil {
				log.Printf("error with parse date: %v\n", err)
			}
			newsInfo.PublishDate = publishDate

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

			s.Find("div.news-item__content > a.news-item__tag ").Each(func(i int, tagSelection *goquery.Selection) {
				tagText := strings.Trim(tagSelection.Text(), " \n\t\r")

				// ====================================================================
				//	get TagID
				// ====================================================================
				tagId, err := repo.Database.GetTagIdByName(tagText)
				if err != nil {
					log.Printf("error with get tag id by name: %v\n", err)
				}

				// ====================================================================
				//	if there is no such tags then create a new and get ID
				// ====================================================================
				if tagId == 0 {
					tagId, err = repo.Database.CreateTags(tagText, ru)
					if err != nil {
						log.Printf("error with create tag by name: %v\n", err)
						return
					}
				}

				_, err = repo.Database.CreateNewsTags(newsId, tagId)
				if err != nil {
					log.Printf("error with create news tags: %v\n", err)
					return
				}
			})

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
	}

	// ====================================================================
	// iterate articles (tm, ru)
	// ====================================================================
	return http.StatusOK, lastLink
}

func getMonthByRussianName(s string) string {
	if strings.Contains(s, "январ") { return "01" }
	if strings.Contains(s, "феврал") { return "02" }
	if strings.Contains(s, "март") { return "03" }
	if strings.Contains(s, "апрел") { return "04" }
	if strings.Contains(s, "мая") { return "05" }
	if strings.Contains(s, "июн") { return "06" }
	if strings.Contains(s, "июл") { return "07" }
	if strings.Contains(s, "август") { return "08" }
	if strings.Contains(s, "сентя") { return "09" }
	if strings.Contains(s, "октя") { return "10" }
	if strings.Contains(s, "ноя") { return "11" }
	if strings.Contains(s, "дека") { return "12" }
	return "impossible"
}