package jvnlist

import (
	"io"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type HeadlineMap map[int][]*Headline

type Headline struct {
	Title string
	Link  string
}

func ParseHeadlines(r io.Reader) (HeadlineMap, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	hm := HeadlineMap{}
	doc.Find(".listbox").Each(func(i int, s *goquery.Selection) {
		ys := s.Find("h2").Text()
		if len(ys) < 4 {
			return
		}
		y, err := strconv.Atoi(ys[:4])
		if err != nil {
			return
		}
		var headlines []*Headline
		s.Find("dl").Each(func(j int, dl *goquery.Selection) {
			title := dl.Find("dt").Text()
			link, exist := dl.Find("a").Attr("href")
			if !exist {
				return
			}
			headlines = append(headlines, &Headline{title, link})
		})
		hm[y] = headlines
	})
	return hm, nil
}
