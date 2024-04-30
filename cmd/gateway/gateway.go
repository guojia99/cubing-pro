/*
 *  * Copyright (c) 2023 guojia99 All rights reserved.
 *  * Created: 2023/7/21 下午9:03.
 *  * Author: guojia(https://github.com/guojia99)
 */

package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/unrolled/secure"
)

func NewCmd() *cobra.Command {
	var pem string
	var key string

	var host string
	var port int
	var httpsPort int
	var indexPath string
	var staticPath string
	var xStaticPath string
	var xFilePath string
	var apiPort int

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "魔方赛事系统网关",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set the router as the default one shipped with Gin
			router := gin.Default()

			// 静态文件
			router.Static("/static", staticPath)
			router.Static("/x-static", xStaticPath)
			router.Static("/x-file", xFilePath)

			// 核心启动器
			router.NoRoute(baseRoute(indexPath, apiPort))

			// Start and run the server
			if pem != "" && key != "" {
				router.Use(tlsHandler(httpsPort, host))
				go router.RunTLS(fmt.Sprintf(":%d", httpsPort), pem, key)
			}

			return router.Run(fmt.Sprintf(":%d", port))
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&host, "host", "localhost", "重定向路径")
	flags.StringVar(&pem, "pem", "", "pem证书路径")
	flags.StringVar(&key, "key", "", "证书key路径")
	flags.IntVarP(&port, "port", "p", 80, "http端口")
	flags.IntVarP(&httpsPort, "https_port", "s", 443, "https端口")
	flags.StringVar(&indexPath, "index", "./build/index.html", "前端启动文件")
	flags.StringVar(&staticPath, "static", "./build/static", "前端静态文件路径")
	flags.StringVar(&xStaticPath, "x-static", "./x-static", "后端静态文件路径")
	flags.StringVar(&xFilePath, "x-file", "./x-file", "其他资源文件路径")
	flags.IntVarP(&apiPort, "api-port", "a", 20000, "后端端口")
	return cmd
}

func baseRoute(indexFile string, apiPort int) gin.HandlerFunc {
	// api
	v2Api, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", apiPort))
	proxyV2 := httputil.NewSingleHostReverseProxy(v2Api)
	return func(ctx *gin.Context) {
		// bot 文件
		if strings.Contains(ctx.Request.URL.Path, ".json") {
			numbers := GetNumbers(ctx.Request.URL.Path)
			if len(numbers) != 0 {
				ctx.JSON(http.StatusOK, gin.H{"bot_appid": int(numbers[0])})
				return
			}
		}

		// api
		if strings.Contains(ctx.Request.URL.Path, "/v2/api") {
			proxyV2.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		// 前端启动器
		ctx.File(indexFile)
	}
}

func tlsHandler(port int, host string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(
			secure.Options{
				SSLRedirect: true,
				SSLHost:     host + ":" + strconv.Itoa(port),
			},
		)
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadGateway, gin.H{
					"error": err,
				},
			)
			return
		}

		c.Next()
	}
}

func GetNumbers(in string) []float64 {
	re := regexp.MustCompile("(-?\\d+)(\\.\\d+)?")
	numbers := re.FindAllString(in, -1)

	var out []float64
	for _, num := range numbers {
		f, err := strconv.ParseFloat(num, 64)
		if err == nil {
			out = append(out, f)
		}
	}
	return out
}

//
//func copyFile(dstName, srcName string) (written int64, err error) {
//	src, err := os.Open(srcName)
//	if err != nil {
//		return
//	}
//	defer src.Close()
//
//	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
//	if err != nil {
//		return
//	}
//	defer dst.Close()
//
//	return io.Copy(dst, src)
//}

//func main() {
//	cmd := NewCmd()
//	if err := cmd.Execute(); err != nil {
//		panic(err)
//	}
//}
