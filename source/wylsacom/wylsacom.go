package wylsacom

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
	ru = ""
)

type Categories struct {
	link string
	name string
	id 	 int
}

var urlParts [2]string
var cat []Categories

func NewsContentParser(repo *repository.Repository, newsText models.NewsText) {
	res, err := http.Get(newsText.Url)
	if err != nil {
		log.Printf("http get %s error: %v\n", newsText.Url, err)
		return
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
		return
	}

	// ====================================================================
	// load the HTML document
	// ====================================================================
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("load html document error: %v\n", err)
		return
	}

	// ====================================================================
	// find title block and get title
	// ====================================================================
	block := doc.Find("main#main > article")

	title := block.Find("section.headline > h1").Text()

	newsText.Title = title

	// ====================================================================
	// newsText to db
	// ====================================================================
	newsTextId, e := repo.Database.CreateNewsText(newsText)

	if e != nil {
		log.Printf("error with create news text %v\n", e)
		return
	}

	// ====================================================================
	// find article text and iterate children tags
	// ====================================================================
	content := block.Find("section.content > div.content__inner")

	content.Children().Each(func(i int, s *goquery.Selection) {
		// ====================================================================
		// get Tag value
		// ====================================================================
		tagValue, err := s.Html()
		tagValue = strings.Trim(tagValue, " ")
		if err != nil {
			log.Printf("error get tag value: %v\n", err)
			return
		}
		// ====================================================================
		// if tag is empty then return
		// ====================================================================
		if len(tagValue) == 0 {
			return
		}
		// ====================================================================
		// skip of links another posts
		// ====================================================================
		if s.Nodes[0].Data == "a" && s.Nodes[0].Attr[0].Val  == "embeded-post" {
			return
		}
		// ====================================================================
		// if text of tag is empty then
		// check and add all possible images on tag to storage
		// ====================================================================
		text := strings.Trim(s.Text(), " \n\t\r")

		if len(text) == 0 || s.Nodes[0].Data == "div" || s.Nodes[0].Data == "figure" {
			var imageLinks []string
			s.Find("img").Each(func(i int, s *goquery.Selection) {
				if attr, ok := s.Attr("src"); !ok {
					return
				} else {
					// check for exists of picture
					for _, v := range imageLinks {
						if attr == v { return }
					}
					imageLinks = append(imageLinks, attr)

					// make attribute
					attr = strings.Trim(attr, " ")

					// make NewsContent
					newsContent := models.NewsContent{
						newsTextId,
						"",
						"img",
						[]models.Attributes{
							models.Attributes{
								Key: "src",
								Value: attr,
							},
						},
					}

					// NewsContent to db
					contentId, contentErr := repo.Database.CreateNewsContent(newsContent)
					if contentErr != nil {
						log.Printf("error with create news content: %v\n", contentErr)
						return
					}

					// Image to storage on "content" bucket
					uploadErr := repo.Storage.UploadImage(context.Background(), "content", attr, strconv.Itoa(contentId))
					if uploadErr != nil {
						log.Printf("error with upload image: %v\n", uploadErr)
					}
				}
			})
			return
		}

		// ====================================================================
		// remove all <img> tag-s on parent tag
		// because there is text inside the tag
		// ====================================================================
		s.Find("img").Each(func(i int, img *goquery.Selection) {
			img.Remove()
		})

		tagValue, err = s.Html()

		if err != nil {
			log.Printf("error get tag value: %v\n", err)
			return
		}
		// ====================================================================
		// make NewsContent
		// analysis attributes of tags. because inside the <a> tag there can be an href attribute that refers to the link
		// ====================================================================
		newsContent := models.NewsContent{
			newsTextId,
			tagValue,
			s.Nodes[0].Data,
			[]models.Attributes{},
		}

		for _, v := range s.Nodes[0].Attr {
			if v.Key == "href" {
				newsContent.Attr = append(newsContent.Attr, models.Attributes{
					Key:   v.Key,
					Value: v.Val,
				})
			}
		}

		// ====================================================================
		// newsContent to db
		// ====================================================================
		_, contentErr := repo.Database.CreateNewsContent(newsContent)
		if contentErr != nil {
			log.Printf("error with create news content: %v\n", contentErr)
			return
		}

	})

	log.Printf("%s) %s parsed\n", newsText.Hl, newsText.Title)
}

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
	sel := doc.Find("div#content > div#primary > main#main > section.postsGrid").Children().Filter("div.postCard")
	for i := range sel.Nodes {
		s := sel.Eq(i)
		// ====================================================================
		// article parse
		// ====================================================================

		// ====================================================================
		// image
		// ====================================================================
		imageLink, ok := s.Find("figure").Attr("style")
		entryPos := strings.Index(imageLink, "https://")
		imageLink = imageLink[entryPos:]
		imageLink = imageLink[:len(imageLink)-1]

		if !ok {
			log.Printf("No news images\n")
			continue
		}
		newsInfo.Image=imageLink

		// ====================================================================
		// link
		// ====================================================================
		link, ok := s.Find("a").Eq(0).Attr("href")
		if !ok {
			log.Printf("No news link\n")
			continue
		}

		// ====================================================================
		// publishDate
		// ====================================================================
		metaSel := s.Find("div.postCard-content > div.postCard-meta > div.postCard-timestamp")
		publishDateStr := strings.Trim(metaSel.Text(), " \n\t\r")

		splitDate := strings.Split(publishDateStr, " ")

		if len(splitDate[2]) <= 4 {
			splitPublishTime := "00:00:00"
			splitDate[1] = getMonthByRussianName(splitDate[1])
			if splitDate[1] == "impossible" {
				continue
			}
			splitPublishDate := splitDate[0]+"."+splitDate[1]+"."+splitDate[2]
			publishDateStr = splitPublishTime + " " + splitPublishDate + " +03:00"
		} else {
			splitPublishTime := splitDate[0]
			timeType := splitDate[1]

			var durationType string
			if strings.Contains(timeType, "??????") { durationType = "h" } else {
				if strings.Contains(timeType, "??????") { durationType = "m" } else {
					if strings.Contains(timeType, "??????") { durationType = "s" } else {
						if strings.Contains(timeType, "????") || strings.Contains(timeType, "??????") {
							cntDays, daysErr := strconv.Atoi(splitPublishTime)
							if daysErr != nil {
								log.Printf("error with calc duration: %v\n", daysErr)
								continue
							}
							splitPublishTime = strconv.Itoa(cntDays * 24)
							durationType = "h"
						}
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
		newsId, e := repo.Database.CreateNews(newsInfo)
		if e != nil {
			log.Printf("Error with create news: %v\n", e)
			continue
		}

		s.Find("div.postCard-meta > div.postCard-tag > object > a").Each(func(i int, tagSelection *goquery.Selection) {
			tagText := strings.Trim(tagSelection.Text(), " \n\t\r")

			// ====================================================================
			//	get TagID
			// ====================================================================
			tagId, err := repo.Database.GetTagIdByName(tagText)
			if err != nil {
				log.Printf("error with get tag id by name: %v\n", err)
			}

			//// ====================================================================
			////	if there is no such tags then create a new and get ID
			//// ====================================================================
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

func StartParser(repo *repository.Repository, newsInfo models.News) {
	cat, err := getCategories(repo)
	if err != nil {
		log.Printf("error with get category id: %v\n", err)
		return
	}

	urlParts[0] = "https://wylsa.com/category/"
	for i := 0; i < categoryCount; i++ {
		urlParts[1] = cat[i].link
		for indexPage := 1; ; indexPage++ {
			// ====================================================================
			// make URL
			// ====================================================================
			newsUrl := urlParts[0] + urlParts[1] + "/page/" + strconv.Itoa(indexPage)

			newsId, err := repo.Database.GetLatestNewsIdByAuthorAndCategory(cat[i].id, newsInfo.AuthID)
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

func getMonthByRussianName(s string) string {
	if strings.Contains(s, "??????????") { return "01" }
	if strings.Contains(s, "????????????") { return "02" }
	if strings.Contains(s, "????????") { return "03" }
	if strings.Contains(s, "??????????") { return "04" }
	if strings.Contains(s, "??????") { return "05" }
	if strings.Contains(s, "??????") { return "06" }
	if strings.Contains(s, "??????") { return "07" }
	if strings.Contains(s, "????????????") { return "08" }
	if strings.Contains(s, "??????????") { return "09" }
	if strings.Contains(s, "????????") { return "10" }
	if strings.Contains(s, "??????") { return "11" }
	if strings.Contains(s, "????????") { return "12" }
	return "impossible"
}

func getCategories(repo *repository.Repository) ([]Categories, error){
	var category []Categories

	category = append(category, Categories{
		link: "news",
		name: "????????????????????",
		id: 0,
	})

	category = append(category, Categories{
		link: "articles",
		name: "????????????????????",
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