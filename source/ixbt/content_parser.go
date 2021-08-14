package ixbt

import (
	"context"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/Hudayberdyyev/crawler/repository"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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
	block := doc.Find("div.b-article")

	title := block.Find("div.b-article__header > h1").Text()

	newsText.Title = title

	// ====================================================================
	// newsText to db
	// ====================================================================
	newsTextId, e := repo.CreateNewsText(newsText)

	if e != nil {
		log.Printf("error with create news text %v\n", e)
		return
	}

	// ====================================================================
	// find article text and iterate children tags
	// ====================================================================
	content := block.Find("div.b-article__content")

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
		// if text of tag is empty then
		// check and add all possible images on tag to storage
		// ====================================================================
		text := strings.Trim(s.Text(), " \n\t\r")

		if len(text) == 0 || s.Nodes[0].Data == "div" {
			var imageLinks []string
			s.Find("img").Each(func(i int, s *goquery.Selection) {
				if attr, ok := s.Attr("src"); !ok {
					return
				} else {
					// check for exists of picture
					for _, v := range imageLinks {
						if attr == v {
							return
						}
					}
					imageLinks = append(imageLinks, attr)

					// make attribute
					attr = strings.Trim(attr, " ")
					attr = "https://www.ixbt.com" + attr
					// ====================================================================
					// if this is first image then update NewsImage
					// ====================================================================
					if cap(imageLinks) == 1 {
						imageLink := attr
						err = repo.Database.UpdateNewsImageById(newsText.NewsID, imageLink)
						if err != nil {
							log.Printf("error with update image by newsId: %v\n", err)
						}
						err = repo.Storage.RemoveImage(context.Background(), "news", strconv.Itoa(newsText.NewsID))
						if err != nil {
							log.Printf("error with remove image by newsId: %v\n", err)
						}

						err = repo.Storage.UploadImage(context.Background(), "news", imageLink, strconv.Itoa(newsText.NewsID))
						if err != nil {
							log.Printf("error with update image: %v\n", err)
						}
					}
					// make NewsContent
					newsContent := models.NewsContent{
						newsTextId,
						"",
						"img",
						[]models.Attributes{
							models.Attributes{
								Key:   "src",
								Value: attr,
							},
						},
					}

					// NewsContent to db
					contentId, contentErr := repo.CreateNewsContent(newsContent)
					if contentErr != nil {
						log.Printf("error with create news content: %v\n", contentErr)
						return
					}

					// Image to storage on "content" bucket
					uploadErr := repo.UploadImage(context.Background(), "content", attr, strconv.Itoa(contentId))
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
		_, contentErr := repo.CreateNewsContent(newsContent)
		if contentErr != nil {
			log.Printf("error with create news content: %v\n", contentErr)
			return
		}

	})

	log.Printf("%s) %s parsed\n", newsText.Hl, newsText.Title)
}
