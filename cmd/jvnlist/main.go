package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/tenkoh/go-jvnlist"
)

const (
	crawlTimeout = 10
	jvnEndpoint  = "https://jvn.jp"
	listPath     = "report/all.html"
)

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
	client := &http.Client{
		Timeout: crawlTimeout * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	headlines, err := jvnlist.ParseHeadlines(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(headlines))
}
