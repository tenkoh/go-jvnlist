package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/tenkoh/go-jvnlist"
)

// todo
// 引数なし: listを更新、差分を取得
// 引数あり：listを更新、指定年を取得

const (
	crawlTimeout  = 10
	crawlInterval = 1
	jvnEndpoint   = "https://jvn.jp"
	listPath      = "report/all.html"
)

var client = &http.Client{
	Timeout: crawlTimeout * time.Second,
}

func getList() ([]*jvnlist.Headline, error) {
	u, err := url.Parse(jvnEndpoint)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, listPath)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	headlines, err := jvnlist.ParseHeadlines(resp.Body)
	if err != nil {
		return nil, err
	}
	return headlines, nil
}

func main() {
	var year int
	flag.IntVar(&year, "y", 0, "specify the year")
	flag.Parse()

	f, err := os.OpenFile("jvnlist.json", os.O_CREATE|os.O_RDWR, 0775)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	recorded := []*jvnlist.Headline{}
	if err := json.NewDecoder(f).Decode(&recorded); err != nil && err != io.EOF {
		log.Fatal(err)
	}

	got, err := getList()
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(f).Encode(got); err != nil {
		panic(err)
	}

	hmap := map[string]struct{}{}
	for _, h := range recorded {
		hmap[h.Link] = struct{}{}
	}
	var added []*jvnlist.Headline
	for _, h := range got {
		_, exist := hmap[h.Link]
		if !exist {
			added = append(added, h)
		}
	}

	// if year is not specified, default action is UPDATE
	if year == 0 {
		if len(added) > 300 {
			log.Fatalln("there are so many added jvn records. please try to specify the year in which you want to get records")
		}
		got = added
	}

	bar := pb.StartNew(len(got))
	defer bar.Finish()
	var details []*jvnlist.Detail
	for _, h := range got {
		bar.Increment()
		if year != 0 {
			if h.PublishedAt.Year() != year {
				continue
			}
		}
		u, err := url.Parse(jvnEndpoint)
		if err != nil {
			log.Printf("error in %s: %s\n", u.String(), err.Error())
			continue
		}
		u.Path = path.Join(u.Path, h.Link)
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			log.Printf("error in %s: %s\n", u.String(), err.Error())
			continue
		}
		time.Sleep(crawlInterval * time.Second)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error in %s: %s\n", u.String(), err.Error())
			continue
		}
		defer resp.Body.Close()
		d, err := jvnlist.ParseDetail(resp.Body)
		if err != nil {
			log.Printf("error in %s: %s\n", u.String(), err.Error())
			continue
		}
		d.Link = u.String()
		details = append(details, d)
	}

	g, err := os.OpenFile("jvn_details.json", os.O_CREATE|os.O_RDWR, 0775)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	ed := []*jvnlist.Detail{}
	if err := json.NewDecoder(g).Decode(&ed); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	dmap := map[string]struct{}{}
	for _, d := range ed {
		dmap[d.Link] = struct{}{}
	}
	for _, d := range details {
		_, exist := dmap[d.Link]
		if !exist {
			ed = append(ed, d)
		}
	}
	if err := json.NewEncoder(g).Encode(ed); err != nil {
		panic(err)
	}
}
