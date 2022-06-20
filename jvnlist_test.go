package jvnlist_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tenkoh/go-jvnlist"
)

const headlines = `<html>
<head></head>
<body>
<div class="listbox">
<h2><a>2022年</a></h2>
<dl>
<dt> 2022/01/01　row1 </dt>
<dd><a href="/link1">text1</a></dd>
</dl>
<dl>
<dt> 2022/01/01　row2 </dt>
<dd><a href="/link2">text2</a></dd>
</dl>
</div>
</body>
</html>
`

func TestParseHeadlines(t *testing.T) {
	got, err := jvnlist.ParseHeadlines(strings.NewReader(headlines))
	if err != nil {
		t.Fatal(err)
	}
	d, _ := time.Parse("2006/01/02", "2022/01/01")
	want := []*jvnlist.Headline{
		{d, "row1", "/link1"},
		{d, "row2", "/link2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %+v, got %+v\n", want, got)
	}
}

func TestParseDetail(t *testing.T) {
	f, err := os.Open(filepath.Clean("./testdata/jvn_sample.html"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := jvnlist.ParseDetail(f)
	if err != nil {
		t.Fatal(err)
	}
	want := &jvnlist.Detail{
		Code:        "JVNVU#000000",
		Title:       "report some problem",
		Abstract:    "There are some problems on a system",
		Target:      "listA listB",
		Detail:      "paragraph listA listB",
		Impact:      "paragraph listA listB",
		Measure:     "update paragraph1 paragraph2 listA listB",
		Vendor:      "vendorA",
		Information: "",
		Supplement:  "",
		Analysis:    "CVSS v3 CVSS:3.0 point: 1.0",
		Reference:   "CVE1 / CVE2",
	}
	if !reflect.DeepEqual(*got, *want) {
		t.Errorf("want %+v, got %+v\n", want, got)
	}
}
