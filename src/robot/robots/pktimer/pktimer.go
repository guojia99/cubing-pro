package pktimer

import (
	"fmt"
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
	end   = "结束" // 强制结束pk
	add   = "加入"
	next  = "下一把" // 强制
)

func (p *PkTimer) WithInPkTimer(msg types.InMessage) bool {
	out, err := p.sendMsgWithOutPkTimer(msg)
	if err != nil {
		log.Print(err)
	}
	return out
}

func (p *PkTimer) sendMsgWithOutPkTimer(msg types.InMessage) (bool, error) {
	if strings.Contains(msg.Message, key) {
		fmt.Println("-------------- in pktimer ")
		return true, p.initPkTimer(msg)
	}
	if !p.checkInPkTimer(msg) {
		return false, nil
	}
	return true, p.runMessage(msg)
}
