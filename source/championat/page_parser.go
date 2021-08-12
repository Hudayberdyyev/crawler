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

func NewsPageParser(repo *repository.Repository, URL string, latestLink string, newsInfo models.News) int {
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
	var ids []int

	fmt.Println(URL)
	sel := doc.Find("div.page-content > div.page-main > div.news._all > div.news-items").Children()
	for i := range sel.Nodes {
		s := sel.Eq(i)
		// ====================================================================
		// if tag is date then get date
		// ====================================================================
		var postDate string
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

			postDate = postDay + "." + postMonth + "." + postYear
			fmt.Println(postDate)
			continue
		}

		continue

		// ====================================================================
		// image
		// ====================================================================
		imageLink, ok := s.Find("img").Attr("src")
		if !ok {
			log.Printf("No news images\n")
			continue
		}
		newsInfo.Image = "https://rozetked.me" + imageLink

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
		metaSel := s.Find("div.post_new-meta > div.post_new-meta-author").Children().Eq(0)
		publishDateStr := strings.Trim(metaSel.Text(), " \n\t\r")

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
			if strings.Contains(timeType, "час") {
				durationType = "h"
			} else {
				if strings.Contains(timeType, "мин") {
					durationType = "m"
				} else {
					if strings.Contains(timeType, "сек") {
						durationType = "s"
					}
				}
			}
			postDuration, err := time.ParseDuration("-" + splitPublishTime + durationType)
			if err != nil {
				log.Printf("error with parse duration %s: %v\n", splitPublishTime, err)
			}
			publishDateStr = time.Now().Add(postDuration).Format(layoutDateTime)
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

		s.Find("div.post_new__main_box_bottom-tags > a").Each(func(i int, tagSelection *goquery.Selection) {
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