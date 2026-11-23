package siteannouncement

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Use the reserved system key namespace so the generic key editor cannot modify it.
const storageKey = "cubing-pro-site-announcements"

var errUnknownAnnouncement = errors.New("公告不存在")

type content struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

type announcement struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
	content
}

type state struct {
	Active bool `json:"active"`
	announcement
}

func read(db *gorm.DB) ([]state, error) {
	result := []state{}
	err := system.GetKeyJSONValue(db, storageKey, &result)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
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
		records, err := read(db.WithContext(c.Request.Context()))
		if err != nil {
			exception.ErrDatabase.ResponseWithError(c, "读取网站公告失败")
			return
		}
		active := make([]announcement, 0, len(records))
		for _, entry := range records {
			if entry.Active {
				active = append(active, entry.announcement)
			}
		}
		exception.ResponseOK(c, active)
	}
}

func AdminGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		records, err := read(db.WithContext(c.Request.Context()))
		respondAdmin(c, records, err)
	}
}

func AdminPublish(db *gorm.DB) gin.HandlerFunc {
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
		id := c.Param("id")
		current, err := change(db.WithContext(c.Request.Context()), id, &input)
		respondAdmin(c, current, err)
	}
}

func AdminDelete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := change(db.WithContext(c.Request.Context()), c.Param("id"), nil)
		respondAdmin(c, current, err)
	}
}

func change(db *gorm.DB, id string, input *content) (state, error) {
	var current state
	err := db.Transaction(func(tx *gorm.DB) error {
		// Initialize the slot safely when two requests publish the first entries
		// together. The write also serializes concurrent updates of this row.
		slot := system.KeyValue{
			StringIDModel: basemodel.StringIDModel{ID: storageKey},
			Value:         "[]",
			Description:   "网站公告",
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&slot).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&slot, "id = ?", storageKey).Error; err != nil {
			return err
		}
		var records []state
		if err := json.Unmarshal([]byte(slot.Value), &records); err != nil {
			return err
		}
		if id == "" {
			newID, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			revision, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			current = state{Active: true, announcement: announcement{
				ID: newID.String(), Revision: revision.String(), content: *input,
			}}
			records = append([]state{current}, records...)
		} else {
			i := 0
			for i < len(records) && records[i].ID != id {
				i++
			}
			if i == len(records) {
				return errUnknownAnnouncement
			}
			current = records[i]
			republish := !current.Active && input != nil
			switch {
			case input == nil:
				if !current.Active {
					return nil
				}
				current.Active = false
			case current.Active && current.content == *input:
				return nil
			default:
				revision, err := uuid.NewRandom()
				if err != nil {
					return err
				}
				current.Active = true
				current.Revision = revision.String()
				current.content = *input
			}
			if republish {
				copy(records[1:i+1], records[:i])
				records[0] = current
			} else {
				records[i] = current
			}
		}
		payload, err := json.Marshal(records)
		if err != nil {
			return err
		}
		return tx.Model(&slot).Update("value", string(payload)).Error
	})
	return current, err
}

func respondAdmin(c *gin.Context, data any, err error) {
	if errors.Is(err, errUnknownAnnouncement) {
		exception.ErrResourceNotFound.ResponseWithError(c, errUnknownAnnouncement.Error())
		return
	}
	if err != nil {
		exception.ErrDatabase.ResponseWithError(c, "读写网站公告失败")
		return
	}
	exception.ResponseOK(c, data)
}
