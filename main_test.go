package main_test

import (
	"reflect"
	"strings"
	"testing"

	jvnlist "github.com/tenkoh/go-jvnlist"
)

const headlines = `<html>
<head></head>
<body>
<div class="listbox">
<h2><a>2022年</a></h2>
<dl>
<dt>row1</dt>
<dd><a href="/link1">text1</a></dd>
</dl>
<dl>
<dt>row2</dt>
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
	wants := []*jvnlist.Headline{
		{"row1", "/link1"},
		{"row2", "/link2"},
	}
	want := jvnlist.HeadlineMap{}
	want[2022] = wants
	if !reflect.DeepEqual(got, want) {
		t.Errorf("not expected result")
		for key, values := range got {
			t.Errorf("%d\n", key)
			for _, v := range values {
				t.Errorf("%+v\n", v)
			}
		}
	}
}
