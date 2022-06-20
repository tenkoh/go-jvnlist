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

const (
	crawlTimeout      = 10
	crawlInterval     = 1
	crawlDefaultLimit = 50
	jvnEndpoint       = "https://jvn.jp"
	listPath          = "report/all.html"
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

func loadRecordedDetails() ([]*jvnlist.Detail, error) {
	f, err := os.OpenFile("jvnlist.json", os.O_CREATE|os.O_RDWR, 0775)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recorded := []*jvnlist.Detail{}
	if err := json.NewDecoder(f).Decode(&recorded); err != nil && err != io.EOF {
		return nil, err
	}
	return recorded, nil
}

func main() {
	limit := crawlDefaultLimit
	flag.IntVar(&limit, "l", crawlDefaultLimit, "set limit num to get details")
	flag.Parse()

	recorded, err := loadRecordedDetails()
	if err != nil {
		log.Fatal(err)
	}

	got, err := getList()
	if err != nil {
		log.Fatal(err)
	}

	hmap := map[string]struct{}{}
	for _, r := range recorded {
		// fit url style to headlines (only path)
		u, err := url.Parse(r.Link)
		if err != nil {
			continue
		}
		hmap[u.Path] = struct{}{}
	}
	updates := make([]*jvnlist.Headline, 0, len(got))
	for _, h := range got {
		_, exist := hmap[path.Clean(h.Link)]
		if !exist {
			updates = append(updates, h)
		}
	}
	log.Printf("[INFO] %d updates are found", len(updates))
	if len(updates) > limit {

		log.Printf("[WARN] updates are more than the limit. %d details are recorded. please try giving an argumen '-l [limit int]' to change the limit", limit)
		updates = updates[:limit]
	}

	bar := pb.StartNew(len(updates))
	defer bar.Finish()
	details := make([]*jvnlist.Detail, len(updates))
	for i, h := range updates {
		bar.Increment()
		u, _ := url.Parse(jvnEndpoint)
		u.Path = path.Join(u.Path, h.Link)
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			log.Printf("[ERROR] %s: %s\n", u.String(), err.Error())
			continue
		}
		time.Sleep(crawlInterval * time.Second)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[ERROR] %s: %s\n", u.String(), err.Error())
			continue
		}
		defer resp.Body.Close()
		d, err := jvnlist.ParseDetail(resp.Body)
		if err != nil {
			log.Printf("[ERROR] %s: %s\n", u.String(), err.Error())
			continue
		}
		d.Link = u.String()
		details[i] = d
	}

	recorded = append(recorded, details...)
	f, err := os.Create("jvnlist.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(recorded); err != nil {
		log.Fatal("[ERROR] fail to save jvn details into a file")
	}
}
