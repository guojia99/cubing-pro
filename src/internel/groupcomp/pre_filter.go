package groupcomp

import (
	"errors"

	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
)

// FilterReview applies the same allowlisted predicates before count and pagination.
func FilterReview(q *gorm.DB, finish, finishDetail, reviewStatus string) (*gorm.DB, error) {
	alias := ""
	switch reviewStatus {
	case "":
	case "passed":
		alias = result.DetailOk
	case "rejected":
		alias = result.DetailNot
	default:
		return nil, errors.New("review_status须为passed/rejected")
	}
	switch finishDetail {
	case "", result.DetailOk, result.DetailNot, result.DetailTimeout:
	default:
		return nil, errors.New("finishDetail须为ok/not/timeout")
	}
	if alias != "" {
		if finishDetail != "" && finishDetail != alias {
			return nil, errors.New("review_status与finishDetail冲突")
		}
		finishDetail = alias
	}
	if finishDetail != "" {
		if finish == "false" {
			return nil, errors.New("终态筛选不能与finish=false同时使用")
		}
		if finish == "" {
			finish = "true"
		}
		q = q.Where("finish = ? AND finish_detail = ?", true, finishDetail)
	}
	switch finish {
	case "", "false":
		q = q.Where("finish = ?", false)
	case "true":
		q = q.Where("finish = ?", true)
	case "all":
	default:
		return nil, errors.New("finish须为true/false/all")
	}
	return q, nil
}
