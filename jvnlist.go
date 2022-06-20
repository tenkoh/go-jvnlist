package jvnlist

import (
	"errors"
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
	Link        string `json:"link"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	Abstract    string `json:"abstract"`
	Target      string `json:"target"`
	Detail      string `json:"detail"`
	Impact      string `json:"impact"`
	Measure     string `json:"measure"`
	Vendor      string `json:"vendor"`
	Information string `json:"information"`
	Supplement  string `json:"supplement"`
	Analysis    string `json:"analysis"`
	Reference   string `json:"reference"`
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
		case "概要":
			detail.Abstract = plained
		case "影響を受けるシステム":
			detail.Target = plained
		case "詳細情報":
			detail.Detail = plained
		case "想定される影響":
			detail.Impact = plained
		case "対策方法":
			detail.Measure = plained
		case "ベンダ情報":
			v, err := parseVendor(s)
			if err != nil {
				return
			}
			detail.Vendor = v
		case "参考情報":
			detail.Information = plained
		case "JPCERT/CCからの補足情報":
			detail.Supplement = plained
		case "JPCERT/CCによる脆弱性分析結果":
			a, _ := parseAnalysis(s)
			detail.Analysis = a
		case "関連文書":
			r, err := parseReference(s)
			if err != nil {
				return
			}
			detail.Reference = r
		}
	})
	return detail, nil
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
	r := strings.NewReplacer("\n", " ", "\t", " ", "　", " ")
	p := r.Replace(s)
	p = TrimSpaceExp.ReplaceAllString(p, " ")
	p = strings.TrimSpace(p)
	return p
}
