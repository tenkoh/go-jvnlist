package jvnlist

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var TrimSpaceExp *regexp.Regexp = regexp.MustCompile(" +")

type Headline struct {
	PublishedAt time.Time `json:"published_at"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
}

type Detail struct {
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Link        string    `json:"link"`
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Abstract    string    `json:"abstract"`
	Target      string    `json:"target"`
	Detail      string    `json:"detail"`
	Impact      string    `json:"impact"`
	Measure     string    `json:"measure"`
	Vendor      string    `json:"vendor"`
	Information string    `json:"information"`
	Supplement  string    `json:"supplement"`
	Analysis    string    `json:"analysis"`
	Reference   string    `json:"reference"`
}

func ParseHeadlines(r io.Reader) ([]*Headline, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var headlines []*Headline
	doc.Find(".listbox").Each(func(i int, s *goquery.Selection) {
		s.Find("dl").Each(func(j int, dl *goquery.Selection) {
			dt := plainText(dl.Find("dt").Text())
			dts := strings.Split(dt, " ")
			if len(dts) != 2 {
				return
			}
			d, err := time.Parse("2006/01/02", dts[0])
			if err != nil {
				return
			}
			title := dts[1]
			link, exist := dl.Find("a").Attr("href")
			if !exist {
				return
			}
			headlines = append(headlines, &Headline{d, title, link})
		})
	})
	return headlines, nil
}

func ParseDetail(r io.Reader) (*Detail, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	detail := new(Detail)
	published, updated, err := parseDate(doc)
	if err != nil {
		return nil, err
	}
	detail.PublishedAt = published
	detail.UpdatedAt = updated
	title, code, err := parseTitle(doc.Find("h1"))
	if err != nil {
		return nil, err
	}
	detail.Code = code
	detail.Title = title
	doc.Find(".textbox").Each(func(i int, s *goquery.Selection) {
		field, exist := s.Find("h2").Find("img").Attr("alt")
		if !exist {
			return
		}
		plained := plainText(s.Text())
		switch field {
		case "??????":
			detail.Abstract = plained
		case "??????????????????????????????":
			detail.Target = plained
		case "????????????":
			detail.Detail = plained
		case "?????????????????????":
			detail.Impact = plained
		case "????????????":
			detail.Measure = plained
		case "???????????????":
			v, err := parseVendor(s)
			if err != nil {
				return
			}
			detail.Vendor = v
		case "????????????":
			detail.Information = plained
		case "JPCERT/CC?????????????????????":
			detail.Supplement = plained
		case "JPCERT/CC??????????????????????????????":
			a, _ := parseAnalysis(s)
			detail.Analysis = a
		case "????????????":
			r, err := parseReference(s)
			if err != nil {
				return
			}
			detail.Reference = r
		}
	})
	return detail, nil
}

func parseDate(doc *goquery.Document) (published, updated time.Time, err error) {
	raw := doc.Find("#head-bar-txt").Text()
	dates := strings.Split(raw, "???")
	if len(dates) != 2 {
		err = errors.New("could not parse published and updated dates")
		return
	}
	published, err = time.Parse("????????????2006/01/02", dates[0])
	if err != nil {
		err = fmt.Errorf("could not parse published date: %w", err)
		return
	}
	updated, err = time.Parse("??????????????????2006/01/02", dates[1])
	if err != nil {
		err = fmt.Errorf("could not parse updated date: %w", err)
		return
	}
	err = nil
	return
}

func parseVendor(s *goquery.Selection) (string, error) {
	tb := s.Find("tbody").First()
	var vs []string
	tb.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			// skip header
			return
		}
		v := strings.TrimSpace(s.Find("td").First().Text())
		if v != "" {
			vs = append(vs, v)
		}
	})
	return strings.Join(vs, " / "), nil
}

func parseReference(s *goquery.Selection) (string, error) {
	tb := s.Find("tbody").First()
	var vs []string
	tb.Find("tr").Each(func(i int, s *goquery.Selection) {
		v := strings.TrimSpace(s.Find("td").First().Next().Text())
		if v != "" {
			vs = append(vs, v)
		}
	})
	return strings.Join(vs, " / "), nil
}

func parseAnalysis(s *goquery.Selection) (string, error) {
	var cs []string
	s.Find(".cvss-line").Each(func(i int, ss *goquery.Selection) {
		cs = append(cs, ss.Text())
	})
	a := strings.Join(cs, " / ")
	return plainText(a), nil
}

func parseTitle(s *goquery.Selection) (title, code string, err error) {
	hs := strings.Split(s.Text(), "\n")
	vs := make([]string, 0, len(hs))
	for _, h := range hs {
		v := strings.TrimSpace(h)
		if len(v) == 0 {
			continue
		}
		vs = append(vs, v)
	}
	if len(vs) < 2 {
		err = errors.New("could not parse detail. not found title or code")
		return
	}
	code = vs[0]
	title = strings.Join(vs[1:], " ")
	return
}

func plainText(s string) string {
	r := strings.NewReplacer("\n", " ", "\t", " ", "???", " ")
	p := r.Replace(s)
	p = TrimSpaceExp.ReplaceAllString(p, " ")
	p = strings.TrimSpace(p)
	return p
}
