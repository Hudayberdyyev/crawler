package repository

import (
	"fmt"
	"github.com/Hudayberdyyev/crawler/models"
	"github.com/jackc/pgx"
	"log"
)

const (
	newsContentTable = "news_content"
	newsTable = "news"
	newsTextTable = "news_text"
	categoryTextTable = "categories_text"
	tagsTextTable = "tags_text"
	tagsTable = "tags"
	newsTagsTable = "news_tags"
)

type NewsDatabase struct {
	db  *pgx.Conn
}

func NewNewsDatabase(db *pgx.Conn) *NewsDatabase {
	return &NewsDatabase{db: db}
}

func (n *NewsDatabase) GetLatestNewsIdByAuthorAndCategory (catId, authId int) (int, error) {
	var id int
	query := fmt.Sprintf("select id from %s where category_id=$1 and author_id=$2 order by publish_date desc limit 1", newsTable)
	err := n.db.QueryRow(query, catId, authId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (n *NewsDatabase) GetLatestNewsUrlByNewsId(newsId int) (string, error) {
	var url string
	query := fmt.Sprintf("select url from %s where news_id=$1 and hl='ru'", newsTextTable)
	err := n.db.QueryRow(query, newsId).Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (n *NewsDatabase) CreateNews(newsInfo models.News) (int, error) {
	var id int

	query := fmt.Sprintf("insert into %s (category_id, author_id, image, publish_date) values ($1, $2, $3, $4) returning id", newsTable)
	err := n.db.QueryRow(query, newsInfo.CatID, newsInfo.AuthID, newsInfo.Image, newsInfo.PublishDate).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (n *NewsDatabase) CreateNewsText(newsText models.NewsText) (int, error) {
	var id int

	if newsText.Hl == "" {
		newsText.Hl = "ru"
	}

	query := fmt.Sprintf("insert into %s (news_id, hl, title, url) values ($1, $2, $3, $4) returning id", newsTextTable)
	err := n.db.QueryRow(query, newsText.NewsID, newsText.Hl, newsText.Title, newsText.Url).Scan(&id)

	if err != nil {
		log.Fatal(err)
	}

	return id, nil
}

func (n *NewsDatabase) CreateNewsContent(newsContent models.NewsContent) (int, error) {
	query := fmt.Sprintf("insert into %s (value, tag, news_text_id, attr) values ($1, $2, $3, $4) returning id", newsContentTable)
	row := n.db.QueryRow(query, newsContent.Value, newsContent.Tag, newsContent.NewsTextID, newsContent.Attr)

	var id int
	err := row.Scan(&id)

	if err != nil {
		log.Fatal(err)
	}

	return id, nil
}

func (n *NewsDatabase) GetCategoryIdByName(categoryName string) (int, error) {
	var id int
	query := fmt.Sprintf("select category_id from %s where title=$1 limit 1", categoryTextTable)
	err := n.db.QueryRow(query, categoryName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (n *NewsDatabase) GetTagIdByName(tagName string) (int, error) {
	var id int
	query := fmt.Sprintf("select tag_id from %s where name=$1 limit 1", tagsTextTable)
	err := n.db.QueryRow(query, tagName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (n *NewsDatabase) CreateTagsText(tagId int, tagName string, hl string) (int, error) {
	var id int
	query := fmt.Sprintf("insert into %s (name, tag_id, hl) values ($1, $2, $3) returning id", tagsTextTable)
	err := n.db.QueryRow(query, tagName, tagId, hl).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (n *NewsDatabase) CreateTags(tagName string, hl string) (int, error) {
	var id int
	query := fmt.Sprintf("insert into %s default values returning id", tagsTable)
	err := n.db.QueryRow(query).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = n.CreateTagsText(id, tagName, hl)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (n *NewsDatabase) CreateNewsTags(newsId int, tagId int) (int, error) {
	var id int
	query := fmt.Sprintf("insert into %s (news_id, tag_id) values ($1, $2) returning id", newsTagsTable)
	err := n.db.QueryRow(query, newsId, tagId).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (n *NewsDatabase) UpdateNewsImageById(newsId int, imageLink string) (error) {
	tx, err := n.db.Begin()
	if err != nil {

		return err
	}
	defer tx.Rollback()
	query := fmt.Sprintf("update %s set image = $1 where id = $2", newsTable)
	_, err = tx.Exec(query, imageLink, newsId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (n *NewsDatabase) UpdateTagByContentId(contentId int, tagName string) (error) {
	tx, err := n.db.Begin()
	if err != nil {

		return err
	}
	defer tx.Rollback()
	query := fmt.Sprintf("update %s set tag = $1 where id = $2", newsContentTable)
	_, err = tx.Exec(query, tagName, contentId)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (n *NewsDatabase) GetNewsIdByUrl(url string) (int, error) {
	var id int
	query := fmt.Sprintf("select news_id from %s where url=$1 limit 1", newsTextTable)
	err := n.db.QueryRow(query, url).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}