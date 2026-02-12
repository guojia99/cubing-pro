package public

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/patrickmn/go-cache"
)

type outputClass struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type AlgorithmGroupsResponse struct {
	CubeKeys []string                 `json:"CubeKeys"`
	ClassMap map[string][]outputClass `json:"ClassMap"`
}

func AlgorithmGroups(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := &AlgorithmGroupsResponse{
			CubeKeys: algs.CubeKeyList,
			ClassMap: make(map[string][]outputClass), // key with cube
		}
		for _, alg := range algs.GetAlgorithms() {
			for _, k := range alg.ClassList {
				o := outputClass{
					Name:  k.Name,
					Image: k.Sets[0].AlgorithmGroups[0].Algorithms[0].Image,
				}
				resp.ClassMap[alg.Cube] = append(resp.ClassMap[alg.Cube], o)
			}
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

type AlgorithmGroupsWithCubeResponse struct {
	algs.AlgorithmClass
}

func AlgorithmGroupsWithCube(svc *svc.Svc) gin.HandlerFunc {
	var cacheData = cache.New(120*time.Minute, 120*time.Minute)
	return func(ctx *gin.Context) {
		cubeID := ctx.Param("cubeID")
		classID := ctx.Param("classID")
		if cubeID == "" || classID == "" {
			exception.ErrRequestBinding.ResponseWithError(ctx, fmt.Errorf("request error"))
			return
		}

		key := fmt.Sprintf("%s-%s", cubeID, classID)
		if data, ok := cacheData.Get(key); ok {
			ctx.JSON(http.StatusOK, data)
			return
		}
		resp := AlgorithmGroupsWithCubeResponse{}

		cube, ok := algs.GetAlgorithms()[cubeID]
		if !ok {
			exception.ErrResourceNotFound.ResponseWithError(ctx, fmt.Errorf("not this cube"))
			return
		}

		for _, v := range cube.ClassList {
			if v.Name == classID {
				resp = AlgorithmGroupsWithCubeResponse{
					AlgorithmClass: v,
				}
				break
			}
		}
		if resp.Name == "" {
			exception.ErrResourceNotFound.ResponseWithError(ctx, fmt.Errorf("not this class"))
			return
		}

		cacheData.Set(key, resp, 120*time.Minute)
		ctx.JSON(http.StatusOK, resp)
	}
}
