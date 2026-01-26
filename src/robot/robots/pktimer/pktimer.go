package pktimer

import (
	"log"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type PkTimer struct {
	Svc *svc.Svc

	SendMessage func(out types.OutMessage) error
}

const (
	key = "pktimer"

	start = "开始" // 开始pk
	out   = "踢"
	exit  = "退出"
	end   = "结束" // 强制结束pk
	add   = "加入"
	next  = "下一把" // 强制

	reload      = "再来一轮"
	reloadStart = "再开一轮"
	update      = "修改"
)

// WithInPkTimer 检查消息是否在PK计时器上下文中，并处理消息
func (p *PkTimer) WithInPkTimer(msg types.InMessage) bool {
	out, err := p.sendMsgWithOutPkTimer(msg)
	if err != nil {
		log.Printf("处理PK计时器消息失败: %v", err)
	}
	return out
}

// sendMsgWithOutPkTimer 处理PK计时器相关消息
func (p *PkTimer) sendMsgWithOutPkTimer(msg types.InMessage) (bool, error) {
	// 检查是否是初始化命令
	if strings.Contains(msg.Message, key) {
		return true, p.initPkTimer(msg)
	}

	// 检查是否在PK上下文中
	if !p.checkInPkTimer(msg) {
		return false, nil
	}

	// 处理PK相关消息
	return true, p.runMessage(msg)
}
