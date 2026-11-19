package advertisements

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"gorm.io/gorm"
)

const storageKey = "__advertisements__"

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

type Product struct {
	ImageURL string `json:"imageUrl"`
	Caption  string `json:"caption"`
	Href     string `json:"href"`
}

type Showcase struct {
	ImageURL string `json:"imageUrl"`
	Caption  string `json:"caption"`
}

type Advertisement struct {
	Key         string     `json:"key"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Contact     string     `json:"contact"`
	ImageURL    string     `json:"imageUrl"`
	TargetURL   string     `json:"targetUrl"`
	Duration    float64    `json:"duration"`
	Enabled     bool       `json:"enabled"`
	Products    []Product  `json:"products"`
	Showcase    []Showcase `json:"showcase"`
}

func validateURL(value string, required bool) error {
	if value == "" && !required {
		return nil
	}
	if value == "" || strings.TrimSpace(value) != value {
		return errors.New("invalid url")
	}
	// Check escaped characters too: browsers may normalize them before navigation.
	decoded := value
	for i := 0; ; i++ {
		if strings.ContainsAny(decoded, `\`) || strings.ContainsFunc(decoded, unicode.IsControl) || strings.HasPrefix(decoded, "//") {
			return errors.New("unsafe url")
		}
		next, err := url.PathUnescape(decoded)
		if err != nil {
			return fmt.Errorf("invalid url escape: %w", err)
		}
		if next == decoded {
			break
		}
		if i == 4 {
			return errors.New("excessively escaped url")
		}
		decoded = next
	}
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if strings.HasPrefix(value, "/") && u.Host == "" && u.Scheme == "" {
		return nil
	}
	if (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" && u.User == nil && u.Opaque == "" {
		return nil
	}
	return errors.New("url must be a site absolute path or http(s) url")
}

func validate(items []Advertisement) error {
	if items == nil || len(items) > 50 {
		return errors.New("expected an array of at most 50 advertisements")
	}
	seen := make(map[string]bool, len(items))
	for i := range items {
		a := &items[i]
		if !keyPattern.MatchString(a.Key) || a.Key == "cuberwenjun" || a.Key == "contact" || seen[a.Key] {
			return fmt.Errorf("advertisement %d: invalid, reserved or duplicate key", i)
		}
		seen[a.Key] = true
		if strings.TrimSpace(a.Title) == "" || utf8.RuneCountInString(a.Title) > 200 || utf8.RuneCountInString(a.Description) > 4000 || utf8.RuneCountInString(a.Contact) > 500 {
			return fmt.Errorf("advertisement %d: invalid title, description or contact length", i)
		}
		if !(a.Duration >= 1 && a.Duration <= 120) || len(a.Products) > 20 || len(a.Showcase) > 20 {
			return fmt.Errorf("advertisement %d: invalid duration or item count", i)
		}
		if err := validateURL(a.ImageURL, true); err != nil {
			return fmt.Errorf("advertisement %d imageUrl: %w", i, err)
		}
		if err := validateURL(a.TargetURL, false); err != nil {
			return fmt.Errorf("advertisement %d targetUrl: %w", i, err)
		}
		for j, p := range a.Products {
			if err := validateURL(p.ImageURL, true); err != nil {
				return fmt.Errorf("advertisement %d product %d imageUrl: %w", i, j, err)
			}
			if err := validateURL(p.Href, false); err != nil {
				return fmt.Errorf("advertisement %d product %d href: %w", i, j, err)
			}
		}
		for j, p := range a.Showcase {
			if err := validateURL(p.ImageURL, true); err != nil {
				return fmt.Errorf("advertisement %d showcase %d imageUrl: %w", i, j, err)
			}
		}
		if a.Products == nil {
			a.Products = []Product{}
		}
		if a.Showcase == nil {
			a.Showcase = []Showcase{}
		}
	}
	return nil
}

func read(db *gorm.DB) ([]Advertisement, error) {
	var items []Advertisement
	err := system.GetKeyJSONValue(db, storageKey, &items)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []Advertisement{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read advertisements: %w", err)
	}
	if err = validate(items); err != nil {
		return nil, fmt.Errorf("invalid stored advertisements: %w", err)
	}
	return items, nil
}

func Get(db *gorm.DB, public bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := read(db.WithContext(c.Request.Context()))
		if err != nil {
			exception.ErrInternalServer.ResponseWithError(c, "读取广告失败")
			return
		}
		if public {
			enabled := make([]Advertisement, 0, len(items))
			for _, item := range items {
				if item.Enabled {
					enabled = append(enabled, item)
				}
			}
			items = enabled
		}
		c.JSON(http.StatusOK, items)
	}
}

// decodeJSON bounds reads even for chunked requests and rejects trailing JSON.
func decodeJSON(c *gin.Context, maxBytes int64, value interface{}) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return err
		}
		return errors.New("expected a single json value")
	}
	return nil
}

func requestError(c *gin.Context, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"message": "请求体过大"})
		return
	}
	exception.ErrValidationFailed.ResponseWithError(c, err.Error())
}

func Put(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []Advertisement
		if err := decodeJSON(c, 8<<20, &items); err != nil {
			requestError(c, err)
			return
		}
		if err := validate(items); err != nil {
			requestError(c, err)
			return
		}
		if err := system.SetKeyJSONValue(db.WithContext(c.Request.Context()), storageKey, items, "广告列表"); err != nil {
			exception.ErrDatabase.ResponseWithError(c, "保存广告失败")
			return
		}
		exception.ResponseOK(c, nil)
	}
}
