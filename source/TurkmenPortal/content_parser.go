package TurkmenPortal

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

func NewsContentParser(repo *repository.Repository, newsInfo models.News, firstNewsId *int, newsText models.NewsText) (int){
	// ====================================================================
	// collect URL
	// ====================================================================
	origUrl := newsText.Url
	URL := newsText.Url

	hl := ""
	if newsText.Hl == tm {
		hl = tm +"/"
	}

	URL = URL[:26] + hl + URL[26:]
	newsText.Url = URL

	// ====================================================================
	// http get URL
	// ====================================================================
	res, err := http.Get(URL)
	if err != nil {
		log.Printf("http get %s error: %v\n", URL, err)
		if strings.Contains(err.Error(), "no such host") {
			return http.StatusRequestTimeout
		}
		return http.StatusGatewayTimeout
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return http.StatusNotFound
	}

	if newsText.Hl != tm {
		//	article to db
		// ====================================================================
		newsId, e := repo.Database.CreateNews(newsInfo)
		if e != nil {
			log.Printf("Error with create news: %v\n", e)
		}
		// ===================================================================
		// if firstNewsId don't set then set firstNewsId to newsId this news
		// ===================================================================
		if *firstNewsId == 0 { *firstNewsId = newsId }
		// image article to storage
		// ====================================================================
		uploadErr := repo.Storage.UploadImage(context.Background(), "news", newsInfo.Image, strconv.Itoa(newsId))
		if uploadErr != nil {
			log.Printf("error with upload image: %v\n", uploadErr)
		}
		newsText.NewsID = newsId
	} else {
		newsId, e := repo.Database.GetNewsIdByUrl(origUrl)
		if e != nil {
			log.Printf("error with get newsId by url: %v\n", e)
			return http.StatusInternalServerError
		}
		newsText.NewsID = newsId
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
	block := doc.Find(".col-sm-9.border-left.level2_cont_right")

	title := block.Find("h1").Eq(0).Text()

	newsText.Title = title

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
	content := block.Find("div.article_text")

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
		if len(s.Text()) == 0 {
			s.Find("img").Each(func(i int, s *goquery.Selection) {
				if attr, ok := s.Attr("src"); !ok {
					return
				} else {
					// make attribute
					attr = strings.Trim(attr, " ")
					if !strings.Contains(attr, "http") { attr = "https://turkmenportal.com" + attr }

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
	return http.StatusOK
}