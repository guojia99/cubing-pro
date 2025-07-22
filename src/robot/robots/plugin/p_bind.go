package plugin

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type BindPlugin struct {
	Svc *svc.Svc
}

func (b *BindPlugin) ID() []string {
	return []string{"bind", "绑定", "登记", "unbind", "解绑", "注销"}
}

func (b *BindPlugin) Help() string {
	return `绑定:
1. 绑定 {你的CubeID} | 绑定QQ ID
2. 绑定 | 获取绑定的信息
3. 解绑 | 解除自己的绑定
4. 解绑 {CubeID|QQ_ID} | 解除用户的绑定(管理员权限)
`
}

func (b *BindPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	if utils.ContainsString(message.Message, "unbind", "解绑", "注销") {
		return b.unbind(message)
	}
	if msg := utils.ReplaceAll(message.Message, "", "bind", "绑定", "登记", " "); len(msg) >= 2 {
		return b.bind(message)
	}
	return b.getUser(message)
}

func (b *BindPlugin) getUser(message types.InMessage) (*types.OutMessage, error) {
	usr, err := getUser(b.Svc, message)
	if err == nil && usr.ID != 0 {
		return message.NewOutMessagef("你的绑定用户为: %s\nCubeID: %s\n主页:https://cubing.pro/x/player/%s", usr.Name, usr.CubeID, usr.CubeID), nil
	}
	return message.NewOutMessage("你还未绑定任何用户"), nil
}

func (b *BindPlugin) bind(message types.InMessage) (*types.OutMessage, error) {
	msg := utils.ReplaceAll(message.Message, "", b.ID()...)
	msg = utils.ReplaceAll(msg, "", " ")
	if len(msg) <= 9 {
		return message.NewOutMessage("请按照 '绑定 2024XXXX01' 的格式来进行绑定"), nil
	}

	if myUser, err := getUser(b.Svc, message); err == nil {
		return message.NewOutMessagef("你已绑定了 %s | %s", myUser.QQUniID, myUser.QQ), nil
	}

	var withCubeIdUser user.User
	if err := b.Svc.DB.Where("cube_id = ?", msg).First(&withCubeIdUser).Error; err != nil || withCubeIdUser.ID == 0 {
		return message.NewOutMessagef("找不到'%s'的用户", msg), nil
	}

	if message.QQBot != "" {
		withCubeIdUser.QQUniID = message.QQBot
	}
	if message.QQ != 0 {
		withCubeIdUser.QQ = fmt.Sprintf("%d", message.QQ)
	}

	if err := b.Svc.DB.Save(&withCubeIdUser).Error; err != nil {
		return message.NewOutMessagef("绑定失败%s", err), nil
	}
	return message.NewOutMessage("绑定成功"), nil
}

func (b *BindPlugin) unbind(message types.InMessage) (*types.OutMessage, error) {
	myUser, err := getUser(b.Svc, message)
	if err != nil {
		return message.NewOutMessage("你未绑定任何帐号"), nil
	}

	// 管理员
	msg := utils.ReplaceAll(utils.ReplaceAll(message.Message, "", b.ID()...), "", " ")
	if len(msg) >= 10 && (myUser.CheckAuth(user.AuthAdmin, user.AuthSuperAdmin)) {
		var withCubeIdUser user.User
		if err = b.Svc.DB.Where("cube_id = ?", msg).First(&withCubeIdUser).Error; err != nil || withCubeIdUser.ID == 0 {
			return message.NewOutMessagef("[管理员] 找不到'%s'的用户", msg), nil
		}

		if withCubeIdUser.QQ == "" && withCubeIdUser.QQUniID == "" {
			return message.NewOutMessagef("[管理员] %s已无绑定", withCubeIdUser.Name), nil
		}

		withCubeIdUser.QQUniID = ""
		withCubeIdUser.QQ = ""
		if err = b.Svc.DB.Save(&withCubeIdUser).Error; err != nil {
			return message.NewOutMessagef("[管理员] 解绑%s失败%s", withCubeIdUser.Name, err), nil
		}
		return message.NewOutMessagef("[管理员] 解绑%s成功", withCubeIdUser.Name), nil
	}

	if message.QQ != 0 {
		myUser.QQ = ""
	}
	if message.QQBot != "" {
		myUser.QQUniID = ""
	}

	if err = b.Svc.DB.Save(&myUser).Error; err != nil {
		return message.NewOutMessagef("解绑失败%s", err), nil
	}
	return message.NewOutMessage("解绑成功"), nil
}
