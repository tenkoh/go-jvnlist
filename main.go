package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	crawlTimeout  = 10
	crawlInterval = 1
	jvnEndpoint   = "https://jvn.jp"
	listPath      = "report/all.html"
)

type ClientWithInterval struct {
	client   *http.Client
	interval int
}

func NewClientWithInterval() ClientWithInterval {
	client := &http.Client{
		Timeout: crawlTimeout * time.Second,
	}
	c := ClientWithInterval{
		interval: crawlInterval,
		client:   client,
	}
	return c
}

func (c *ClientWithInterval) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	time.Sleep(time.Duration(c.interval) * time.Second)
	return resp, err
}

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

func main() {
	u, err := url.Parse(jvnEndpoint)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, listPath)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}
	client := NewClientWithInterval()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	headlines, err := ParseHeadlines(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(headlines))
}
