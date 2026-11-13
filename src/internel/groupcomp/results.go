package groupcomp

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ValidateValues(values []float64, route event.RouteType) error {
	m := route.RouteMap()
	if m.Rounds < 1 || len(values) == 0 || len(values) > m.Rounds {
		return errors.New("成绩数量与赛制不符")
	}
	if m.Repeatedly && len(values)%3 != 0 {
		return errors.New("多盲成绩必须为成功数、尝试数、时间三元组")
	}
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("成绩必须是有限数值")
		}
		if m.Repeatedly && i%3 != 2 {
			if v < 0 || v != math.Trunc(v) {
				return errors.New("多盲数量必须是非负整数")
			}
			if i%3 == 1 && (values[i-1] > v || (v == 0 && values[i+1] != result.DNS)) {
				return errors.New("多盲成功数/尝试数不合法")
			}
			continue
		}
		if v <= 0 && v != result.DNF && v != result.DNS && v != result.DNP && v != result.DNT {
			return errors.New("无效成绩值")
		}
	}
	return nil
}

// Lock the parent even when no result row exists, serializing all competition writers.
func Lock(db *gorm.DB, id uint) (competition.Competition, error) {
	var c competition.Competition
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&c, "id = ?", id).Error
	return c, err
}

func Validate(c competition.Competition, eventID string, round interface{}, userID uint, _ time.Time, pending bool, _ bool) (competition.CompetitionEvent, competition.Schedule, error) {
	if !c.IsGroupCompetition() {
		return competition.CompetitionEvent{}, competition.Schedule{}, errors.New("群赛校验不可用于其他比赛")
	}
	if c.Status != competition.Running {
		return competition.CompetitionEvent{}, competition.Schedule{}, errors.New("比赛未审批")
	}
	if pending && !c.CanPreResult {
		return competition.CompetitionEvent{}, competition.Schedule{}, errors.New("比赛不允许预录入")
	}
	ev, ok := c.EventMap()[eventID]
	if !ok {
		return ev, competition.Schedule{}, errors.New("比赛项目不存在")
	}
	s, err := ev.CurRunningSchedule(round, nil)
	if err != nil {
		return ev, s, err
	}
	if s.RoundNum < 1 {
		return ev, s, errors.New("轮次编号无效")
	}
	if s.RoundNum > 1 && !s.NoRestrictions && !slices.Contains(s.AdvancedToThisRound, userID) {
		return ev, s, errors.New("选手不在晋级名单中")
	}
	return ev, s, nil
}

func CheckPlayer(db *gorm.DB, c competition.Competition, s competition.Schedule, eventID string, u user.User, create bool) error {
	var reg competition.Registration
	err := db.First(&reg, "comp_id = ? AND user_id = ?", c.ID, u.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if c.Genre == competition.OnlineInformal && create {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now()
			reg = competition.Registration{CompID: c.ID, CompName: c.Name, UserID: u.ID, UserName: u.Name, Status: competition.RegisterStatusPass, RegistrationTime: now, AcceptationTime: &now}
		}
		if reg.Status != competition.RegisterStatusPass {
			return errors.New("选手报名未通过审批")
		}
		reg.SetEvent(eventID)
		if err := db.Save(&reg).Error; err != nil {
			return err
		}
	} else if err != nil || reg.Status != competition.RegisterStatusPass || !slices.Contains(reg.EventsList(), eventID) {
		return errors.New("选手未报名项目或报名未通过审批")
	}
	if s.RoundNum > 1 {
		var n int64
		if err := db.Model(&result.Results{}).Where("comp_id = ? AND user_id = ? AND event_id = ? AND round_number = ?", c.ID, u.ID, eventID, s.RoundNum-1).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return errors.New("上轮无成绩无法录入")
		}
	}
	return nil
}

func SavePending(db *gorm.DB, pres []result.PreResults) error {
	if len(pres) == 0 {
		return errors.New("没有成绩")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, pres[0].CompetitionID)
		if err != nil {
			return err
		}
		if !c.IsGroupCompetition() {
			return errors.New("群赛预成绩写入不可用于其他比赛")
		}
		for i := range pres {
			p := &pres[i]
			if p.CompetitionID != c.ID {
				return errors.New("批量成绩必须属于同一比赛")
			}
			ev, s, err := Validate(c, p.EventID, p.RoundNumber, p.UserID, time.Now(), true, true)
			if err != nil {
				return err
			}
			p.EventRoute, p.EventName, p.Round, p.RoundName = ev.EventRoute, ev.EventName, s.Round, s.Round
			if err = ValidateValues(p.Result, ev.EventRoute); err != nil {
				return err
			}
			var u user.User
			if err = tx.First(&u, "id = ?", p.UserID).Error; err != nil {
				return err
			}
			p.CubeID, p.PersonName = u.CubeID, u.Name
			p.CompetitionName, p.CompsName = c.Name, c.Name
			if err = CheckPlayer(tx, c, s, p.EventID, u, true); err != nil {
				return err
			}
			var old result.PreResults
			err = tx.Where("comp_id = ? AND user_id = ? AND event_id = ? AND round_number = ? AND finish = ?", c.ID, p.UserID, p.EventID, p.RoundNumber, false).Order("id ASC").First(&old).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				p.Model = old.Model
			}
			p.Finish, p.Detail, p.FinishDetail, p.Processor, p.ProcessorID, p.ResultID = false, "", "", "", 0, nil
			p.Result = result.UpdateOrgResult(p.Result, p.EventRoute, s.Cutoff, s.CutoffNumber, s.TimeLimit)
			if err = p.Update(); err != nil {
				return err
			}
			if err = tx.Save(p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func AddResult(tx *gorm.DB, c competition.Competition, u user.User, eventID string, round int, values []float64, penalty result.Penalty, at time.Time, open bool) (result.Results, error) {
	if !c.IsGroupCompetition() {
		return LegacyAddResult(tx, c, u, eventID, round, values, penalty)
	}
	return addResult(tx, c, u, eventID, round, values, penalty, at, open, nil)
}

// Approval still verifies the persisted record's identity and event/round metadata.
func addResult(tx *gorm.DB, c competition.Competition, u user.User, eventID string, round int, values []float64, penalty result.Penalty, at time.Time, open bool, historical *result.PreResults) (result.Results, error) {
	var ev competition.CompetitionEvent
	var s competition.Schedule
	var err error
	if historical == nil {
		ev, s, err = Validate(c, eventID, round, u.ID, at, false, open)
	} else {
		ev, s, err = validatePreResult(c, *historical)
	}
	if err != nil {
		return result.Results{}, err
	}
	if err = ValidateValues(values, ev.EventRoute); err != nil {
		return result.Results{}, err
	}
	if err = CheckPlayer(tx, c, s, eventID, u, true); err != nil {
		return result.Results{}, err
	}
	var r result.Results
	err = tx.Where("comp_id = ? AND user_id = ? AND event_id = ? AND round_number = ?", c.ID, u.ID, eventID, round).First(&r).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return r, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r = result.Results{CompetitionID: c.ID, CompetitionName: c.Name, UserID: u.ID, PersonName: u.Name, CubeID: u.CubeID, EventID: ev.EventID, EventName: ev.EventName, EventRoute: ev.EventRoute, Round: s.Round, RoundNumber: s.RoundNum}
	}
	r.Result = result.UpdateOrgResult(values, ev.EventRoute, s.Cutoff, s.CutoffNumber, s.TimeLimit)
	r.Penalty = penalty
	r.PenaltyJSON = ""
	if err = r.Update(); err != nil {
		return r, err
	}
	err = tx.Save(&r).Error
	return r, err
}

func Approve(db *gorm.DB, compID, preID uint, detail string, processor user.User) error {
	return approve(db, compID, preID, detail, processor, false, "")
}

// AdminApprove is only wired to the admin route, with role checks also enforced here.
func AdminApprove(db *gorm.DB, compID, preID uint, detail string, processor user.User, historicalConfirmed bool, reviewReason string) error {
	if !processor.CheckAuth(user.AuthAdmin, user.AuthSuperAdmin) || processor.ID == 0 || processor.Ban {
		return errors.New("历史人工核查审批仅限有效管理员或超级管理员")
	}
	return approve(db, compID, preID, detail, processor, historicalConfirmed, reviewReason)
}

func approve(db *gorm.DB, compID, preID uint, detail string, processor user.User, confirmed bool, reason string) error {
	reason = strings.TrimSpace(reason)
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, compID)
		if err != nil {
			return err
		}
		var p result.PreResults
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, "id = ? AND comp_id = ?", preID, c.ID).Error; err != nil {
			return err
		}
		if p.Finish {
			return errors.New("成绩已处理")
		}
		if c.Status != competition.Running {
			return errors.New("比赛未审批")
		}
		if !c.IsGroupCompetition() {
			if confirmed {
				return errors.New("历史人工确认仅限群赛")
			}
			return legacyApprove(tx, c, p, detail, processor)
		}
		if !slices.Contains([]string{result.DetailOk, result.DetailNot, result.DetailWait, result.DetailDoubtful, result.DetailTimeout}, detail) {
			return errors.New("无效审批状态")
		}
		if p.Detail == detail && (detail == result.DetailWait || detail == result.DetailDoubtful) {
			return nil
		}
		if detail == result.DetailOk {
			var u user.User
			if err = tx.First(&u, "id = ? AND cube_id = ?", p.UserID, p.CubeID).Error; err != nil {
				return fmt.Errorf("历史预成绩选手不匹配: %w", err)
			}
			r, err := addResult(tx, c, u, p.EventID, p.RoundNumber, p.Result, p.Penalty, p.CreatedAt, false, &p)
			if err != nil {
				return err
			}
			p.ResultID = &r.ID
		}
		previousDetail := p.Detail
		p.Processor, p.ProcessorID, p.Detail = processor.Name, processor.ID, detail
		if confirmed {
			p.Detail = fmt.Sprintf("%s; HistoricalConfirmed=true; ReviewReason=%q", detail, reason)
			if previousDetail != "" {
				p.Detail += fmt.Sprintf("; PreviousDetail=%q", previousDetail)
			}
		}
		if detail == result.DetailOk || detail == result.DetailNot || detail == result.DetailTimeout {
			p.Finish, p.FinishDetail = true, detail
		}
		// updated_at remains the last content submission, not the later review time.
		op := tx.Model(&result.PreResults{}).Where("id = ? AND comp_id = ? AND finish = ?", p.ID, c.ID, false).UpdateColumns(map[string]interface{}{
			"processor": p.Processor, "processor_id": p.ProcessorID, "detail": p.Detail, "finish": p.Finish, "finish_detail": p.FinishDetail, "result_id": p.ResultID,
		})
		if op.Error != nil {
			return op.Error
		}
		if op.RowsAffected != 1 {
			return errors.New("成绩审批状态发生变化")
		}
		return nil
	})
}

func validatePreResult(c competition.Competition, p result.PreResults) (competition.CompetitionEvent, competition.Schedule, error) {
	var ev competition.CompetitionEvent
	var s competition.Schedule
	if p.CompetitionID != c.ID || p.UserID == 0 || p.CubeID == "" || p.RoundNumber < 1 {
		return ev, s, errors.New("预成绩缺少必要归属字段，无法审批通过")
	}
	var err error
	ev, s, err = Validate(c, p.EventID, p.RoundNumber, p.UserID, time.Time{}, true, false)
	if err != nil {
		return ev, s, err
	}
	if p.EventRoute != ev.EventRoute || p.Round != s.Round {
		return ev, s, errors.New("历史预成绩赛制或轮次名称与比赛不匹配")
	}
	return ev, s, nil
}
