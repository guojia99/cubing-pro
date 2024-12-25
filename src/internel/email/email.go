package email

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gopkg.in/gomail.v2"
)

// todo 并发, 考虑后面也要发，是否需要存数据库
func SendEmail(config svc.EmailConfig, subject string, emails []string, message []byte) error {
	var msgs []*gomail.Message
	for _, email := range emails {
		m := gomail.NewMessage()
		m.SetHeader("From", fmt.Sprintf("%s<%s>", config.FromName, config.From))
		m.SetHeader("To", email)
		m.SetHeader("Subject", subject)
		m.SetHeader("Date", time.Now().Format(time.DateTime))
		m.SetHeader("Organization", "CubingPro")
		m.SetBody("text/html", string(message))

		msgs = append(msgs, m)
	}

	d := gomail.NewDialer(config.SmtpHost, config.SmtpPort, config.From, config.Password)

	if err := d.DialAndSend(msgs...); err != nil {
		return err
	}
	return nil
}

func SendEmailWithTemp(config svc.EmailConfig, subject string, email []string, templateStr string, data interface{}) error {
	// 创建template对象
	t := template.New("cubingPro")

	// 解析模板内容
	parse, err := t.Parse(templateStr)
	if err != nil {
		return err
	}

	// 创建一个缓冲区，用于存储渲染后的模板内容
	var tplBuffer bytes.Buffer

	// 将数据填充到模板中
	err = parse.Execute(&tplBuffer, data)
	if err != nil {
		return err
	}

	return SendEmail(config, subject, email, tplBuffer.Bytes())
}
