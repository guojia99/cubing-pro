package user

import basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"

type UserKVType = int

const (
	UserKVTypeString UserKVType = iota + 1
	UserKVTypeInt
	UserKVTypeJSON
	UserKVTypeYaml
	UserKVTypeMD
)

// UserKV 用户kv数据
type UserKV struct {
	basemodel.Model

	UserId uint   `gorm:"uniqueIndex:idx_user_key"` // 联合唯一索引名：idx_user_key
	Key    string `gorm:"uniqueIndex:idx_user_key"` // 相同索引名，表示是同一个组合索引

	Value string     `gorm:"column:value"`
	Type  UserKVType `gorm:"column:type"`
}

const MaxKVLength = 1024 * 1024 * 100 // 100MB

var WhitelistKeys = []string{
	"blind_tightening_assistant", // 盲拧助手
}

func IsInWhitelist(key string) bool {
	for _, v := range WhitelistKeys {
		if v == key {
			return true
		}
	}
	return false
}

/*
CE	R' F R:[S',R U' R']	R' F R S' R U' R' S R U R2 F' R
EC	R' F R:[R U' R',S']	R' F R2 U' R' S' R U R' S R' F' R
CF	U:[S',R U' R']	U S' R U' R' S R U R' U'
FC	U:[R U' R',S']	U R U' R' S' R U R' S U'
CG	R2 U R U R' U' R' U' R' U R'
GC	R U' R U R U R U' R' U' R2
CH	[S,R' F R]	S R' F R S' R' F' R
HC	[R' F R,S]	R' F R S R' F' R S'
DE	U' M U:[M',U2]	U' M U M' U2 M U M' U
ED	U' M U':[M',U2]	U' M U' M' U2 M U' M' U
DF	U': [S,R' F R]	U' S R' F R S' R' F' R U
FD	U':[R' F R,S]	U' R' F R S R' F' R S' U
DG	[r U' r',S']	r U' r' S' r U r' S
GD	[S',r U' r']	S' r U' r' S r U r'
DH	M U:[M',U2]	M U M' U2 M U M'
HD	M U':[M',U2]	M U' M' U2 M U' M'
EG	R2 U':[S,R2]	R2 U' S R2 S' R2 U R2
GE	R2 U':[R2,S]	R2 U' R2 S R2 S' U R2
EH	U M U:[M',U2]	U M U M' U2 M U M' U'
HE	U M U':[M',U2]	U M U' M' U2 M U' M' U'
FG	R':[U' R U,M]	R' U' R U M U' R' U r
GF	R':[M,U' R U]	r' U' R U M' U' R' U R
FH	R:[M',U R' U']	r U R' U' M U R U' R'
HF	R:[U R' U',M']	R U R' U' M' U R U' r'
*/
