package email

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEmailCodeTempDataParser(t *testing.T) {

	tp, err := template.ParseFiles("./base_email_msg.gohtml")
	if err != nil {
		t.Fatal(err)
	}

	var tplBuffer bytes.Buffer

	data := CodeTempData{
		Subject:        "测试主题",
		UserName:       "guojia",
		Option:         "测试操作",
		OptionsTimeOut: time.Now().Format(time.DateTime),
		OptionsCode:    "123456",
		OptionsUrl:     "http://cubing.pro",
		Notify:         "通知",
		NotifyMsg:      strings.Repeat("这是一条非常特别的通知", 10),
		NotifyUrl:      "http://cubing.pro",
	}
	if err = tp.Execute(&tplBuffer, data); err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("./parser_file.html", tplBuffer.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
