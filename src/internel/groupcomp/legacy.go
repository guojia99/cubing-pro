package groupcomp

import (
	"errors"
	"slices"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/gorm"
)

// LegacyAddResult preserves HEAD business decisions for non-group competitions.
func LegacyAddResult(db *gorm.DB, c competition.Competition, u user.User, eventID string, round int, values []float64, penalty result.Penalty) (result.Results, error) {
	var r result.Results
	if c.IsGroupCompetition() {
		return r, errors.New("群赛不能进入原线上录入分支")
	}
	if !c.IsRunningTime() {
		return r, errors.New("不在比赛时间")
	}
	ev, ok := c.EventMap()[eventID]
	if !ok {
		return r, errors.New("比赛项目不存在")
	}
	s, err := ev.CurRunningSchedule(round, nil)
	if err != nil {
		return r, err
	}
	var reg competition.Registration
	err = db.First(&reg, "comp_id = ? AND user_id = ?", c.ID, u.ID).Error
	if c.Genre == competition.OnlineInformal {
		if err != nil {
			now := time.Now()
			reg = competition.Registration{CompID: c.ID, CompName: c.Name, UserID: u.ID, UserName: u.Name, Status: competition.RegisterStatusPass, RegistrationTime: now, AcceptationTime: &now}
		}
		reg.SetEvent(eventID)
		if err = db.Save(&reg).Error; err != nil {
			return r, err
		}
	} else {
		if err != nil {
			return r, err
		}
		if !slices.Contains(reg.EventsList(), eventID) {
			return r, errors.New("该选手未报名该项目")
		}
		if reg.Status != competition.RegisterStatusPass {
			return r, errors.New("该选手比赛资格未审核")
		}
		if !s.FirstRound && !slices.Contains(s.AdvancedToThisRound, u.ID) {
			return r, errors.New("不在晋级名单中")
		}
	}
	if s.RoundNum != 1 {
		var previous result.Results
		if err = db.First(&previous, "user_id = ? AND comp_id = ? AND event_id = ? AND round_number = ?", u.ID, c.ID, eventID, s.RoundNum-1).Error; err != nil {
			return r, errors.New("上轮无成绩无法录入")
		}
	}
	err = db.First(&r, "user_id = ? AND comp_id = ? AND event_id = ? AND round_number = ?", u.ID, c.ID, eventID, s.RoundNum).Error
	if err != nil || r.ID == 0 {
		r = result.Results{CompetitionID: c.ID, CompetitionName: c.Name, Round: s.Round, RoundNumber: s.RoundNum, PersonName: u.Name, UserID: u.ID, CubeID: u.CubeID, EventID: ev.EventID, EventName: ev.EventName, EventRoute: ev.EventRoute}
	}
	r.Result = result.UpdateOrgResult(values, ev.EventRoute, s.Cutoff, s.CutoffNumber, s.TimeLimit)
	r.Penalty = penalty
	if err = r.Update(); err != nil {
		return r, err
	}
	err = db.Save(&r).Error
	return r, err
}

func LegacySavePending(db *gorm.DB, c competition.Competition, u user.User, eventID, round string, values []float64, penalty result.Penalty) (result.PreResults, error) {
	var p result.PreResults
	if c.IsGroupCompetition() {
		return p, errors.New("群赛不能进入原线上预成绩分支")
	}
	var reg competition.Registration
	if err := db.First(&reg, "comp_id = ? AND user_id = ?", c.ID, u.ID).Error; err != nil {
		return p, errors.New("你未报名该场比赛")
	}
	if !slices.Contains(reg.EventsList(), eventID) {
		return p, errors.New("你未报名该项目")
	}
	if reg.Status != competition.RegisterStatusPass {
		return p, errors.New("你的比赛资格未被审核")
	}
	if !c.IsRunningTime() {
		return p, errors.New("不在比赛时间内")
	}
	if !c.CanPreResult {
		return p, errors.New("本场比赛不允许预录入成绩")
	}
	ev, ok := c.EventMap()[eventID]
	if !ok {
		return p, errors.New("本场比赛未开放该项目")
	}
	s, err := ev.CurRunningSchedule(round, nil)
	if err != nil {
		return p, err
	}
	if !s.FirstRound && !s.NoRestrictions && !slices.Contains(s.AdvancedToThisRound, u.ID) {
		return p, errors.New("你不在晋级名单中")
	}
	if err = db.First(&p, "comp_id = ? AND round = ? AND user_id = ? AND event_id = ?", c.ID, round, u.ID, eventID).Error; err == nil && p.Finish {
		p = result.PreResults{}
	}
	model := p.Model
	p = result.PreResults{Results: result.Results{Model: model, CompetitionID: c.ID, CompetitionName: c.Name, Round: s.Round, RoundNumber: s.RoundNum, PersonName: u.Name, UserID: u.ID, CubeID: u.CubeID, Result: values, Penalty: penalty, EventID: ev.EventID, EventName: ev.EventName, EventRoute: ev.EventRoute}, CompsName: c.Name, RoundName: s.Round, Recorder: u.Name, Source: "web-api"}
	if err = p.Update(); err != nil {
		return p, err
	}
	// Persist the actual model rather than HEAD's request DTO; no new qualification rules.
	err = db.Save(&p).Error
	return p, err
}

func legacyApprove(db *gorm.DB, c competition.Competition, p result.PreResults, detail string, processor user.User) error {
	if !c.IsRunningTime() {
		return errors.New("不在比赛时间")
	}
	p.Processor, p.ProcessorID, p.Detail = processor.Name, processor.ID, detail
	if detail == result.DetailOk || detail == result.DetailNot {
		p.Finish, p.FinishDetail = true, detail
	}
	if detail == result.DetailOk {
		var u user.User
		if err := db.First(&u, "cube_id = ?", p.CubeID).Error; err != nil {
			return err
		}
		r, err := LegacyAddResult(db, c, u, p.EventID, p.RoundNumber, p.Result, p.Penalty)
		if err != nil {
			return err
		}
		p.ResultID = &r.ID
	}
	return db.Save(&p).Error
}

func legacyRefresh(db *gorm.DB, c competition.Competition, eventID string, round int, open bool) error {
	ev, ok := c.EventMap()[eventID]
	if !ok {
		return errors.New("不存在该项目")
	}
	s, err := ev.CurRunningSchedule(round, nil)
	if err != nil {
		return err
	}
	if s.IsRunning == open {
		return nil
	}
	if open {
		run := true
		if last, err := ev.CurRunningSchedule(round, &run); err == nil && last.IsRunning {
			return errors.New("轮次已经开启")
		}
		s.IsRunning = true
		ev.UpdateSchedule(round, s)
	} else {
		s.IsRunning = false
		ev.UpdateSchedule(round, s)
		if next, err := ev.CurRunningSchedule(round+1, nil); err == nil {
			var rs []result.Results
			if err = db.Where("comp_id = ? AND round_number = ? AND event_id = ?", c.ID, round, eventID).Find(&rs).Error; err != nil {
				return err
			}
			result.SortResult(rs)
			next.AdvancedToThisRound = make([]uint, 0)
			for _, r := range rs {
				if next.Competitors == 0 || r.Rank < next.Competitors {
					next.AdvancedToThisRound = append(next.AdvancedToThisRound, r.UserID)
				}
			}
			ev.UpdateSchedule(round+1, next)
		}
	}
	c.UpdateEvent(ev)
	return db.Save(&c).Error
}
