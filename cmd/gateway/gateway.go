/*
 *  * Copyright (c) 2023 guojia99 All rights reserved.
 *  * Created: 2023/7/21 下午9:03.
 *  * Author: guojia(https://github.com/guojia99)
 */

package gateway

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/gateway"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/spf13/cobra"
)

func NewCmd(svc **svc.Svc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "魔方赛事系统网关",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(*svc)
			//fmt.Println(svc.Cfg)
			gw := gateway.NewGateway(*svc)
			return gw.Run()
		},
	}
	return cmd
}
