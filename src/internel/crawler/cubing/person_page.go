/*
 * 粗饼选手主页 HTML 抓取与解析（服务端出站，避免浏览器 CORS）。
 * 不在此包内做并发锁或节流（由 API 层控制）；FetchPersonPage 内 recover，避免 panic 冒泡。
 */

package cubing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	cubingPersonURLFmt = "https://cubing.com/results/person/%s"

	fetchTimeout = 12 * time.Second
	maxBodyBytes = 2 << 20 // 2 MiB
)

// WCA ID：4 位年份 + 4 位大写字母 + 2 位数字（与 WCA 规则一致）
var wcaIDRegexp = regexp.MustCompile(`^[0-9]{4}[A-Z]{4}[0-9]{2}$`)

// PersonPageCode 业务码（与 HTTP 状态分离，便于前端统一处理）
type PersonPageCode string

const (
	PersonCodeOK             PersonPageCode = "OK"
	PersonCodeInvalidWcaID   PersonPageCode = "INVALID_WCA_ID"
	PersonCodeNotFound       PersonPageCode = "NOT_FOUND"
	PersonCodeWcaIDMismatch  PersonPageCode = "WCA_ID_MISMATCH"
	PersonCodeUpstreamStatus PersonPageCode = "UPSTREAM_HTTP_ERROR"
	PersonCodeUpstreamBody   PersonPageCode = "UPSTREAM_READ_ERROR"
	PersonCodeParseError     PersonPageCode = "PARSE_ERROR"
	PersonCodeRecoveredPanic PersonPageCode = "RECOVERED_PANIC"
)

// PersonPageResult 单次抓取结果（含失败原因，便于调用方展示）
type PersonPageResult struct {
	Code    PersonPageCode `json:"code"`
	Message string         `json:"message,omitempty"`

	RequestedWcaID string        `json:"requested_wca_id"`
	Person         *PersonFields `json:"person,omitempty"`
}

// PersonFields 从粗饼页面解析出的选手信息
type PersonFields struct {
	WcaID     string            `json:"wca_id"`
	Name      string            `json:"name"`
	AvatarURL string            `json:"avatar_url,omitempty"`
	Details   map[string]string `json:"details,omitempty"` // 如 地区、参赛次数、性别 等
}

// ValidateWcaIDFormat 校验 WCA ID 字符串格式（不发起网络请求）
func ValidateWcaIDFormat(wcaID string) bool {
	s := strings.TrimSpace(strings.ToUpper(wcaID))
	return len(s) == 10 && wcaIDRegexp.MatchString(s)
}

// FetchPersonPage 请求粗饼选手主页并解析。请使用 context 控制超时；内部 panic 会转为 RECOVERED_PANIC，不向外 panic。
func FetchPersonPage(ctx context.Context, rawWcaID string) (out PersonPageResult) {
	reqID := strings.TrimSpace(strings.ToUpper(rawWcaID))
	out = PersonPageResult{
		RequestedWcaID: reqID,
		Code:           PersonCodeInvalidWcaID,
		Message:        "WCA ID 格式无效（应为 10 位：4 数字 + 4 大写字母 + 2 数字）",
	}
	if !ValidateWcaIDFormat(reqID) {
		return out
	}

	defer func() {
		if r := recover(); r != nil {
			out = PersonPageResult{
				RequestedWcaID: reqID,
				Code:           PersonCodeRecoveredPanic,
				Message:        fmt.Sprintf("内部异常: %v", r),
			}
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(cubingPersonURLFmt, reqID), nil)
	if err != nil {
		out.Code = PersonCodeUpstreamBody
		out.Message = err.Error()
		return out
	}
	req.Header.Set("User-Agent", "cubing-pro-crawler/1.0 (+server-side; cubing.com person sync)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		out.Code = PersonCodeUpstreamBody
		out.Message = err.Error()
		return out
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out.Code = PersonCodeUpstreamStatus
		out.Message = fmt.Sprintf("粗饼 HTTP %d", resp.StatusCode)
		return out
	}

	limited := io.LimitReader(resp.Body, maxBodyBytes)
	html, err := io.ReadAll(limited)
	if err != nil {
		out.Code = PersonCodeUpstreamBody
		out.Message = err.Error()
		return out
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		out.Code = PersonCodeParseError
		out.Message = err.Error()
		return out
	}

	return parseCubingPersonDocument(reqID, doc)
}

// parseCubingPersonDocument 从已加载的粗饼选手页 DOM 解析（供单测，不发起网络请求）
func parseCubingPersonDocument(reqID string, doc *goquery.Document) PersonPageResult {
	out := PersonPageResult{RequestedWcaID: reqID}

	personRoot := doc.Find("div.results-person[data-person-id]").First()
	if personRoot.Length() == 0 {
		out.Code = PersonCodeNotFound
		out.Message = "粗饼上未找到该选手（或页面已变更）"
		return out
	}

	pageWcaID, _ := personRoot.Attr("data-person-id")
	pageWcaID = strings.TrimSpace(strings.ToUpper(pageWcaID))
	if pageWcaID == "" {
		out.Code = PersonCodeParseError
		out.Message = "缺少 data-person-id"
		return out
	}
	if pageWcaID != reqID {
		out.Code = PersonCodeWcaIDMismatch
		out.Message = fmt.Sprintf("页面选手 ID (%s) 与请求 (%s) 不一致", pageWcaID, reqID)
		out.Person = &PersonFields{WcaID: pageWcaID}
		return out
	}

	name := strings.TrimSpace(personRoot.Find("h1.text-center").First().Text())
	avatarURL, _ := personRoot.Find("img.user-avatar").First().Attr("src")
	avatarURL = strings.TrimSpace(avatarURL)

	details := parsePersonDetails(personRoot)

	out.Code = PersonCodeOK
	out.Message = ""
	out.Person = &PersonFields{
		WcaID:     pageWcaID,
		Name:      name,
		AvatarURL: avatarURL,
		Details:   details,
	}
	return out
}

func parsePersonDetails(personRoot *goquery.Selection) map[string]string {
	m := make(map[string]string)
	personRoot.Find(".person-detail .mt-10").Each(func(_ int, s *goquery.Selection) {
		title := strings.TrimSpace(strings.TrimSuffix(s.Find(".info-title").Text(), ":"))
		value := strings.TrimSpace(s.Find(".info-value").Text())
		if title != "" && value != "" {
			m[title] = value
		}
	})
	if len(m) == 0 {
		return nil
	}
	return m
}
