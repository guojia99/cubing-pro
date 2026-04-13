package cubing

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestValidateWcaIDFormat(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"2018GUOZ01", true},
		{"2018guoz01", true},
		{"INVALID", false},
		{"2018GUOZ0", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := ValidateWcaIDFormat(tc.id); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.id, got, tc.want)
		}
	}
}

func TestParseCubingPersonDocument_ok(t *testing.T) {
	html := `<div class="col-lg-12 results-person" data-person-id="2018GUOZ01">
  <h1 class="text-center">Zejia Guo (郭泽嘉)</h1>
  <div class="text-center"><img class="user-avatar" src="https://i.cubing.com/upload/x.jpg" alt="" /></div>
  <div class="panel panel-info person-detail">
    <div class="panel-body">
      <div class="row">
        <div class="col-md-4 mt-10"><span class="info-title">姓名:</span><span class="info-value">Zejia Guo</span></div>
        <div class="col-md-4 mt-10"><span class="info-title">参赛次数:</span><span class="info-value">34</span></div>
      </div>
    </div>
  </div>
</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	res := parseCubingPersonDocument("2018GUOZ01", doc)
	if res.Code != PersonCodeOK {
		t.Fatalf("code %s %s", res.Code, res.Message)
	}
	if res.Person == nil || res.Person.Name != "Zejia Guo (郭泽嘉)" {
		t.Fatalf("name: %+v", res.Person)
	}
	if res.Person.AvatarURL != "https://i.cubing.com/upload/x.jpg" {
		t.Fatalf("avatar: %s", res.Person.AvatarURL)
	}
	if res.Person.Details["参赛次数"] != "34" {
		t.Fatalf("details: %+v", res.Person.Details)
	}
}

func TestParseCubingPersonDocument_notFound(t *testing.T) {
	html := `<html><body><h1>选手</h1></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}
	res := parseCubingPersonDocument("2018GUOZ01", doc)
	if res.Code != PersonCodeNotFound {
		t.Fatalf("want NOT_FOUND got %s", res.Code)
	}
}
