package middleware

import (
	"encoding/json"
	"errors"
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

func GetJwtUser(ctx *gin.Context) (JwtMapClaims, error) {
	mp := jwt.ExtractClaims(ctx)
	val, ok := mp[IdentityKey]
	if !ok {
		return JwtMapClaims{}, errors.New("无权限")
	}
	dataStr, _ := json.Marshal(val)
	var payload JwtMapClaims
	_ = json.Unmarshal(dataStr, &payload)
	return payload, nil
}

func GetAuthUser(ctx *gin.Context) (user2.User, error) {
	val, ok := ctx.Get(authUserKey)
	if !ok {
		err := fmt.Errorf("找不到用户")
		exception.ErrAuthField.ResponseWithError(ctx, err)
		return user2.User{}, err
	}
	return val.(user2.User), nil
}
