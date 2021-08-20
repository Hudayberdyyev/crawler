package championat

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

func NewsContentParser(repo *repository.Repository, newsText models.NewsText) int{
	res, err := http.Get(newsText.Url)
	if err != nil {
		log.Printf("http get %s error: %v\n", newsText.Url, err)
		if strings.Contains(err.Error(), "no such host") {
			return http.StatusRequestTimeout
		}
		return http.StatusGatewayTimeout
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return http.StatusNotFound
	}

	// ====================================================================
	// load the HTML document
	// ====================================================================
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("load html document error: %v\n", err)
		return http.StatusInternalServerError
	}

	// ====================================================================
	// find title block and get title
	// ====================================================================
	block := doc.Find("div.page > div.page-content > div.page-main > article")

	title := block.Find("header > div.article-head__title").Eq(0).Text()

	newsText.Title = title

	// ====================================================================
	// if article has a head_photo then update news image
	// ====================================================================
	if imageLink, ok := block.Find("header > div.article-head__photo > img").Eq(0).Attr("src"); ok {

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

	// ====================================================================
	// newsText to db
	// ====================================================================
	newsTextId, e := repo.Database.CreateNewsText(newsText)

	if e != nil {
		log.Printf("error with create news text %v\n", e)
		return http.StatusInternalServerError
	}

	// ====================================================================
	// find article text and iterate children tags
	// ====================================================================
	content := block.Find("div.article-content")

	content.Children().Each(func(i int, s *goquery.Selection) {
		// ====================================================================
		// get Tag value
		// ====================================================================
		tagValue, err := s.Html()
		tagValue = strings.Trim(tagValue, " \n\t\r")
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
			// ====================================================================
			// skip all external links
			// ====================================================================
			for _, class := range s.Nodes[0].Attr {
				if strings.Contains(class.Val, "external") || strings.Contains(class.Val, "banner") ||
					strings.Contains(class.Val, "match-embed") {
					return
				}
			}
			// ====================================================================
			// find all images and add to repository
			// ====================================================================
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
	return http.StatusOK
}