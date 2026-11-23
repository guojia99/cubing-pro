package advertisements

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/testutil"
	"gorm.io/gorm"
)

// This fixture deliberately avoids svc/NewConvenient and their shared migrations.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.OpenMySQL(t)
	if err := db.AutoMigrate(&system.KeyValue{}, &system.Image{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func validAd() Advertisement {
	return Advertisement{Key: "shop_1-A", Title: "广告", ImageURL: "/static/cover.png", Duration: 3.5, Enabled: true}
}

func request(t *testing.T, handler gin.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

func TestValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Advertisement)
	}{
		{"reserved contact", func(a *Advertisement) { a.Key = "contact" }},
		{"reserved cuberwenjun", func(a *Advertisement) { a.Key = "cuberwenjun" }},
		{"empty key", func(a *Advertisement) { a.Key = "" }},
		{"long key", func(a *Advertisement) { a.Key = strings.Repeat("a", 65) }},
		{"invalid key", func(a *Advertisement) { a.Key = "a/b" }},
		{"blank title", func(a *Advertisement) { a.Title = " \t" }},
		{"long title", func(a *Advertisement) { a.Title = strings.Repeat("中", 201) }},
		{"long description", func(a *Advertisement) { a.Description = strings.Repeat("中", 4001) }},
		{"long contact", func(a *Advertisement) { a.Contact = strings.Repeat("中", 501) }},
		{"missing cover", func(a *Advertisement) { a.ImageURL = "" }},
		{"script target", func(a *Advertisement) { a.TargetURL = "javascript:alert(1)" }},
		{"duration zero", func(a *Advertisement) { a.Duration = 0 }},
		{"duration large", func(a *Advertisement) { a.Duration = 120.1 }},
		{"products overflow", func(a *Advertisement) { a.Products = make([]Product, 21) }},
		{"showcase overflow", func(a *Advertisement) { a.Showcase = make([]Showcase, 21) }},
		{"product missing image", func(a *Advertisement) { a.Products = []Product{{}} }},
		{"product unsafe href", func(a *Advertisement) { a.Products = []Product{{ImageURL: "/p.png", Href: "//evil.test"}} }},
		{"showcase unsafe image", func(a *Advertisement) { a.Showcase = []Showcase{{ImageURL: "data:image/png;base64,x"}} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := validAd()
			tt.edit(&a)
			if err := validate([]Advertisement{a}); err == nil {
				t.Fatal("invalid advertisement accepted")
			}
		})
	}
	a := validAd()
	a.Title, a.Description, a.Contact = strings.Repeat("中", 200), strings.Repeat("中", 4000), strings.Repeat("中", 500)
	a.Products = []Product{{ImageURL: "https://example.com/a.png"}}
	a.Showcase = []Showcase{{ImageURL: "/s.png"}}
	for _, duration := range []float64{1, 1.5, 120} {
		a.Duration = duration
		if err := validate([]Advertisement{a}); err != nil {
			t.Fatal(err)
		}
	}
	if validate(nil) == nil || validate(make([]Advertisement, 51)) == nil || validate([]Advertisement{a, a}) == nil {
		t.Fatal("null, oversized or duplicate list accepted")
	}
	items := make([]Advertisement, 50)
	for i := range items {
		items[i] = validAd()
		items[i].Key = uuid.NewString()
		for j := 0; j < 20; j++ {
			items[i].Products = append(items[i].Products, Product{ImageURL: "/p.png"})
			items[i].Showcase = append(items[i].Showcase, Showcase{ImageURL: "/s.png"})
		}
	}
	if err := validate(items); err != nil {
		t.Fatalf("maximum counts rejected: %v", err)
	}
}

func TestURLValidation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "relative/path", "//evil.test", `/\evil.test`, "/%5cevil", "/%255cevil", "/%2fevil", "/%0aevil", "/a\x00b", " https://example.com", "https://example.com\n", "javascript:alert(1)", "data:text/html,x", "ftp://example.com", "https:///x", "https://user@example.com", "/%xx"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if validateURL(value, true) == nil {
				t.Fatalf("unsafe url accepted: %q", value)
			}
		})
	}
	for _, value := range []string{"/", "/v3/cube-api/static/image/id", "/a%20b.png", "https://example.com/a?q=1#p", "http://localhost:8000/a"} {
		if err := validateURL(value, true); err != nil {
			t.Errorf("valid url %q rejected: %v", value, err)
		}
	}
	if err := validateURL("", false); err != nil {
		t.Fatal(err)
	}
}

func TestHandlersAndStorage(t *testing.T) {
	db := testDB(t)
	for _, public := range []bool{true, false} {
		w := request(t, Get(db, public), http.MethodGet, "")
		if w.Code != 200 || w.Body.String() != "[]" {
			t.Fatalf("missing KV: %d %s", w.Code, w.Body)
		}
	}
	a, hidden := validAd(), validAd()
	a.Products = []Product{{ImageURL: "/p.png", Caption: "商品", Href: "https://example.com/shop"}}
	a.Showcase = []Showcase{{ImageURL: "/s.png", Caption: "展示"}}
	hidden.Key, hidden.Enabled = "hidden", false
	body, err := json.Marshal([]Advertisement{a, hidden})
	if err != nil {
		t.Fatal(err)
	}
	w := request(t, Put(db), http.MethodPut, string(body))
	if w.Code != 200 || w.Body.String() != `{"code":"200","data":null,"msg":"ok"}` {
		t.Fatalf("put: %d %s", w.Code, w.Body)
	}
	for _, public := range []bool{false, true} {
		w = request(t, Get(db, public), http.MethodGet, "")
		var items []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
			t.Fatal(err)
		}
		want := 2
		if public {
			want = 1
		}
		if w.Code != 200 || len(items) != want || len(items[0]) != 10 || items[0]["imageUrl"] != a.ImageURL || items[0]["duration"] != 3.5 {
			t.Fatalf("get public=%v: %d %s", public, w.Code, w.Body)
		}
		if _, ok := items[0]["products"].([]interface{}); !ok {
			t.Fatal("products must be an array")
		}
		if _, ok := items[0]["showcase"].([]interface{}); !ok {
			t.Fatal("showcase must be an array")
		}
		product := items[0]["products"].([]interface{})[0].(map[string]interface{})
		showcase := items[0]["showcase"].([]interface{})[0].(map[string]interface{})
		if len(product) != 3 || product["imageUrl"] != "/p.png" || product["href"] != "https://example.com/shop" || len(showcase) != 2 || showcase["imageUrl"] != "/s.png" {
			t.Fatalf("nested JSON contract: %s", w.Body)
		}
	}
	for _, body := range []string{"null", "{}", "[{}]", "[] []", `[{"unexpected":1}]`} {
		if w = request(t, Put(db), http.MethodPut, body); w.Code != 422 {
			t.Fatalf("bad put %s: %d", body, w.Code)
		}
	}
	if items, err := read(db); err != nil || len(items) != 2 {
		t.Fatalf("invalid put changed storage: %v %v", items, err)
	}
	if w = request(t, Put(db), http.MethodPut, "[]"); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w = request(t, Get(db, true), http.MethodGet, ""); w.Body.String() != "[]" {
		t.Fatal(w.Body)
	}
	if err := system.SetKeyJSONValue(db, storageKey, []Advertisement{hidden}, ""); err != nil {
		t.Fatal(err)
	}
	if w = request(t, Get(db, true), http.MethodGet, ""); w.Code != 200 || w.Body.String() != "[]" {
		t.Fatalf("disabled advertisement leaked: %d %s", w.Code, w.Body)
	}
	for _, corrupt := range []string{"null", "[{}]", "invalid json"} {
		if err := db.Model(&system.KeyValue{}).Where("id = ?", storageKey).Update("value", corrupt).Error; err != nil {
			t.Fatal(err)
		}
		if w = request(t, Get(db, true), http.MethodGet, ""); w.Code != 500 {
			t.Fatalf("corrupt KV %s was swallowed: %d", corrupt, w.Code)
		}
	}
	if err := db.Migrator().DropTable(&system.KeyValue{}); err != nil {
		t.Fatal(err)
	}
	if w = request(t, Get(db, false), http.MethodGet, ""); w.Code != 500 {
		t.Fatalf("read DB error swallowed: %d", w.Code)
	}
	if w = request(t, Put(db), http.MethodPut, "[]"); w.Code != 500 {
		t.Fatalf("write DB error swallowed: %d", w.Code)
	}
}

func TestKVReadFailureDoesNotWrite(t *testing.T) {
	db := testDB(t)
	if err := system.SetKeyJSONValue(db, storageKey, []Advertisement{validAd()}, ""); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("forced query failure")
	if err := db.Callback().Query().Before("gorm:query").Register("fail_kv_query", func(tx *gorm.DB) {
		tx.AddError(failure)
	}); err != nil {
		t.Fatal(err)
	}
	if err := system.SetKeyJSONValue(db, storageKey, []Advertisement{}, ""); !errors.Is(err, failure) {
		t.Fatalf("query failure not propagated: %v", err)
	}
	if err := db.Callback().Query().Remove("fail_kv_query"); err != nil {
		t.Fatal(err)
	}
	items, err := read(db)
	if err != nil || len(items) != 1 {
		t.Fatalf("query failure changed KV: %v %v", items, err)
	}
}

func encodedImage(t *testing.T, format string) string {
	t.Helper()
	var b bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	var err error
	switch format {
	case "png":
		err = png.Encode(&b, img)
	case "jpeg":
		err = jpeg.Encode(&b, img, nil)
	case "gif":
		err = gif.Encode(&b, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	raw := b.Bytes()
	switch format {
	case "jpeg":
		// APP1 contains little-endian TIFF with one SHORT Orientation tag set to 6.
		exif := []byte{
			0xff, 0xe1, 0x00, 0x22,
			'E', 'x', 'i', 'f', 0, 0,
			'I', 'I', 0x2a, 0, 8, 0, 0, 0,
			1, 0, 0x12, 0x01, 3, 0, 1, 0, 0, 0,
			6, 0, 0, 0, 0, 0, 0, 0,
		}
		raw = append(append(append([]byte{}, raw[:2]...), exif...), raw[2:]...)
	case "png":
		// Insert a valid gAMA chunk after IHDR to detect loss of color metadata.
		gamma := make([]byte, 16)
		binary.BigEndian.PutUint32(gamma[:4], 4)
		copy(gamma[4:8], "gAMA")
		binary.BigEndian.PutUint32(gamma[8:12], 45455)
		binary.BigEndian.PutUint32(gamma[12:], crc32.ChecksumIEEE(gamma[4:12]))
		raw = append(append(append([]byte{}, raw[:33]...), gamma...), raw[33:]...)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestImageValidation(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"jpeg", "png"} {
		img, err := decodeImage(encodedImage(t, format))
		if err != nil || img.format != format || img.width != 3 || img.height != 2 {
			t.Fatalf("%s: %s %v", format, img.format, err)
		}
	}
	raw, err := base64.StdEncoding.DecodeString(encodedImage(t, "png"))
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]uint32{{8193, 2}, {5000, 5000}} {
		oversized := bytes.Clone(raw)
		binary.BigEndian.PutUint32(oversized[16:20], size[0])
		binary.BigEndian.PutUint32(oversized[20:24], size[1])
		binary.BigEndian.PutUint32(oversized[29:33], crc32.ChecksumIEEE(oversized[12:29]))
		if _, err := decodeImage(base64.StdEncoding.EncodeToString(oversized)); err == nil {
			t.Fatal("oversized dimensions accepted")
		}
	}
	for _, data := range []string{"", "bad base64", "data:image/png;base64," + encodedImage(t, "png"), encodedImage(t, "gif"), base64.StdEncoding.EncodeToString(raw[:33]), base64.StdEncoding.EncodeToString(make([]byte, maxImageBytes+1))} {
		if _, err := decodeImage(data); err == nil {
			t.Fatal("invalid image accepted")
		}
	}
	// A valid image padded to the byte boundary is accepted, one byte more is not.
	padded := make([]byte, maxImageBytes)
	copy(padded, raw)
	if _, err := decodeImage(base64.StdEncoding.EncodeToString(padded)); err != nil {
		t.Fatalf("5 MiB boundary rejected: %v", err)
	}
}

func TestUploadAndCleanup(t *testing.T) {
	db := testDB(t)
	root := t.TempDir()
	auth := func(*gin.Context) (uint, error) { return 42, nil }
	for _, format := range []string{"jpeg", "png"} {
		data := encodedImage(t, format)
		original, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(ImageRequest{ImageName: "../../untrusted.gif", Data: data})
		if err != nil {
			t.Fatal(err)
		}
		w := request(t, UploadImage(db, root, auth), http.MethodPost, string(body))
		var response struct {
			Code string        `json:"code"`
			Data ImageResponse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || response.Code != "200" || response.Data.Width != 3 || response.Data.Height != 2 || response.Data.URL != "/v3/cube-api/static/image/"+response.Data.UID {
			t.Fatalf("upload: %d %s", w.Code, w.Body)
		}
		if _, err := uuid.Parse(response.Data.UID); err != nil {
			t.Fatal(err)
		}
		var record system.Image
		if err := db.First(&record, "uid = ?", response.Data.UID).Error; err != nil {
			t.Fatal(err)
		}
		wantPath := filepath.Join(root, "advertisements", response.Data.UID+"."+format)
		if record.Use != "advertisement" || record.UserID != 42 || record.LocalPath != wantPath || record.URL != "" {
			t.Fatalf("wrong record: %+v", record)
		}
		saved, err := os.ReadFile(wantPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(saved, original) {
			t.Fatalf("%s upload bytes changed during save", format)
		}
		if format == "jpeg" {
			if !bytes.Equal(saved[6:12], []byte("Exif\x00\x00")) || binary.LittleEndian.Uint16(saved[30:32]) != 6 {
				t.Fatal("EXIF Orientation=6 not preserved")
			}
		} else if !bytes.Equal(saved[37:41], []byte("gAMA")) || binary.BigEndian.Uint32(saved[41:45]) != 45455 {
			t.Fatal("PNG gamma metadata not preserved")
		}
		config, savedFormat, decodeErr := image.DecodeConfig(bytes.NewReader(saved))
		if decodeErr != nil || config.Width != 3 || config.Height != 2 || savedFormat != format {
			t.Fatalf("saved image: %v %s", decodeErr, savedFormat)
		}
	}
	before, err := os.ReadDir(filepath.Join(root, "advertisements"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Create().Before("gorm:create").Register("fail_image_insert", func(tx *gorm.DB) {
		tx.AddError(errors.New("forced image DB failure"))
	}); err != nil {
		t.Fatal(err)
	}
	img, err := decodeImage(encodedImage(t, "png"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := saveImage(db, root, 42, img); err == nil {
		t.Fatal("DB failure ignored")
	}
	body, err := json.Marshal(ImageRequest{Data: encodedImage(t, "png")})
	if err != nil {
		t.Fatal(err)
	}
	if w := request(t, UploadImage(db, root, auth), http.MethodPost, string(body)); w.Code != 500 {
		t.Fatalf("DB failure upload: %d %s", w.Code, w.Body)
	}
	after, err := os.ReadDir(filepath.Join(root, "advertisements"))
	if err != nil || len(before) != len(after) {
		t.Fatalf("failed upload left file: %v %v", after, err)
	}
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := saveImage(db, blocked, 42, img); err == nil {
		t.Fatal("filesystem failure ignored")
	}
}

func TestUploadBodyLimit(t *testing.T) {
	root := t.TempDir()
	auth := func(*gin.Context) (uint, error) { return 1, nil }
	body := `{"ImageName":"x","Data":"` + strings.Repeat("A", int(maxImageRequestBytes)) + `"}`
	for _, lengthKnown := range []bool{true, false} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		if !lengthKnown {
			c.Request.ContentLength = -1
			c.Request.TransferEncoding = []string{"chunked"}
		}
		UploadImage(nil, root, auth)(c)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("oversized body (length known=%v): %d %s", lengthKnown, w.Code, w.Body)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("oversized request wrote files: %v %v", entries, err)
	}
	for _, body := range []string{`{"Data":"bad"}`, `{"Data":""}`, "{}", "null", "{} {}"} {
		w := request(t, UploadImage(nil, root, auth), http.MethodPost, body)
		if w.Code != 422 {
			t.Fatalf("invalid upload: %d %s", w.Code, w.Body)
		}
	}
	denied := func(c *gin.Context) (uint, error) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return 0, errors.New("unauthorized")
	}
	if w := request(t, UploadImage(nil, root, denied), http.MethodPost, "{}"); w.Code != 401 {
		t.Fatalf("unauthorized upload: %d", w.Code)
	}
}

func TestImageDefaultRoot(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	db := testDB(t)
	img, err := decodeImage(encodedImage(t, "png"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := saveImage(db, "", 1, img)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "advertisements", out.UID+".png")); err != nil {
		t.Fatalf("empty staticPath did not use working directory: %v", err)
	}
}
