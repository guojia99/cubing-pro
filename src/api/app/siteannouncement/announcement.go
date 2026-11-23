package siteannouncement

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"gorm.io/gorm"
)

// Use the reserved system key namespace so the generic key editor cannot modify it.
const storageKey = "cubing-pro-site-announcement"

// Serialize changes to the single slot, including comparison before saving.
var writeMu sync.Mutex

type content struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

type announcement struct {
	Revision string `json:"revision"`
	content
}

type state struct {
	Active bool `json:"active"`
	announcement
}

func read(db *gorm.DB) (state, error) {
	var result state
	err := system.GetKeyJSONValue(db, storageKey, &result)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return state{}, nil
	}
	return result, err
}

func (input *content) normalize() error {
	input.Text = strings.TrimSpace(input.Text)
	input.Href = strings.TrimSpace(input.Href)
	if n := utf8.RuneCountInString(input.Text); n < 1 || n > 300 {
		return errors.New("公告正文须为 1 至 300 个字符")
	}
	if !validLink(input.Href) {
		return errors.New("详情链接须为站内绝对路径或 HTTPS 链接")
	}
	return nil
}

func validLink(value string) bool {
	if value == "" {
		return true
	}
	u, err := url.Parse(value)
	if err != nil {
		return false
	}
	// Browsers normalize backslashes and control characters when resolving links.
	if strings.Contains(value, `\`) || strings.ContainsFunc(value, unicode.IsControl) {
		return false
	}
	if strings.HasPrefix(value, "/") {
		return !strings.HasPrefix(value, "//") && u.Host == "" && u.Scheme == ""
	}
	return u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}

func PublicGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		current, err := read(db.WithContext(c.Request.Context()))
		if err != nil {
			exception.ErrDatabase.ResponseWithError(c, "读取网站公告失败")
			return
		}
		if !current.Active {
			exception.ResponseOK(c, nil)
			return
		}
		exception.ResponseOK(c, current.announcement)
	}
}

func AdminGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		current, err := read(db.WithContext(c.Request.Context()))
		respondAdmin(c, current, err)
	}
}

func AdminPut(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input content
		if err := c.ShouldBindJSON(&input); err != nil {
			exception.ErrRequestBinding.ResponseWithError(c, "无法读取公告内容")
			return
		}
		if err := input.normalize(); err != nil {
			exception.ErrValidationFailed.ResponseWithError(c, err.Error())
			return
		}
		current, err := change(db.WithContext(c.Request.Context()), &input)
		respondAdmin(c, current, err)
	}
}

func AdminDelete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := change(db.WithContext(c.Request.Context()), nil)
		respondAdmin(c, current, err)
	}
}

func change(db *gorm.DB, input *content) (state, error) {
	writeMu.Lock()
	defer writeMu.Unlock()

	current, err := read(db)
	if err != nil {
		return state{}, err
	}
	if input == nil {
		if !current.Active {
			return current, nil
		}
		current.Active = false
	} else {
		if current.Active && current.content == *input {
			return current, nil
		}
		revision, err := uuid.NewRandom()
		if err != nil {
			return state{}, err
		}
		current = state{
			Active: true,
			announcement: announcement{
				Revision: revision.String(),
				content:  *input,
			},
		}
	}
	err = system.SetKeyJSONValue(db, storageKey, current, "网站公告")
	return current, err
}

func respondAdmin(c *gin.Context, current state, err error) {
	if err != nil {
		exception.ErrDatabase.ResponseWithError(c, "读写网站公告失败")
		return
	}
	exception.ResponseOK(c, current)
}
