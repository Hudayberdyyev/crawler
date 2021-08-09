package models

import "time"

type News struct {
	CatID 		int
	AuthID 		int
	Image 		string
	PublishDate time.Time
}

type NewsText struct {
	NewsID 		int
	Hl 			string
	Title 		string
	Url 		string
}

type Attributes struct {
	Key 		string `json:"key"`
	Value 		string `json:"value"`
}

type NewsContent struct {
	NewsTextID 	int			`db:"news_text_id"`
	Value 		string		`db:"value"`
	Tag			string			`db:"tag"`
	Attr 		[]Attributes
}
