package public

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
)

type outputClass struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Alg       string `json:"alg"`
	CaseCount int    `json:"caseCount"`
}

type AlgorithmGroupsResponse struct {
	CubeKeys []string                 `json:"CubeKeys"`
	ClassMap map[string][]outputClass `json:"ClassMap"`
}

const catalogOverviewCache = "formula-catalog-overview"

func AlgorithmGroups(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if resp, ok := svc.Cache.Get(catalogOverviewCache); ok {
			ctx.JSON(http.StatusOK, resp)
			return
		}
		var rows []struct {
			Cube      string
			Name      string
			Alg       string
			CaseCount int
		}
		// Choose a preview from the earliest non-empty case, preferring a
		// rotation-free solution in that case. Count cases independently of their solutions.
		err := svc.DB.WithContext(ctx.Request.Context()).Raw(`
SELECT cl.cube, cl.name,
  (SELECT COUNT(*) FROM formula_sets s
    JOIN formula_groups g ON g.set_id = s.id
    JOIN formula_cases c ON c.group_id = g.id WHERE s.class_id = cl.id) AS case_count,
  COALESCE((SELECT sol.moves FROM formula_sets s
    JOIN formula_groups g ON g.set_id = s.id
    JOIN formula_cases c ON c.group_id = g.id
    JOIN formula_solutions sol ON sol.case_id = c.id
    WHERE s.class_id = cl.id AND TRIM(sol.moves) <> ''
    ORDER BY s.position, g.position, c.position,
      CASE WHEN LENGTH(sol.moves) = LENGTH(REPLACE(REPLACE(REPLACE(sol.moves,'x',''),'y',''),'z','')) THEN 0 ELSE 1 END,
      sol.position LIMIT 1), '') AS alg
FROM formula_classes cl ORDER BY cl.cube, cl.position`).Scan(&rows).Error
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		resp := AlgorithmGroupsResponse{CubeKeys: algs.CubeKeyList, ClassMap: make(map[string][]outputClass)}
		for _, row := range rows {
			resp.ClassMap[row.Cube] = append(resp.ClassMap[row.Cube], outputClass{Name: row.Name, Alg: row.Alg, CaseCount: row.CaseCount})
		}
		svc.Cache.Set(catalogOverviewCache, resp, time.Minute)
		ctx.JSON(http.StatusOK, resp)
	}
}

type AlgorithmGroupsWithCubeResponse struct {
	algs.AlgorithmClass
}

func AlgorithmGroupsWithCube(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cubeID, classID := ctx.Param("cubeID"), ctx.Param("classID")
		if cubeID == "" || classID == "" {
			exception.ErrRequestBinding.ResponseWithError(ctx, fmt.Errorf("request error"))
			return
		}
		key := "formula-catalog-class:" + cubeID + "/" + classID
		if resp, ok := svc.Cache.Get(key); ok {
			ctx.JSON(http.StatusOK, resp)
			return
		}
		var class *algs.AlgorithmClass
		err := svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var err error
			class, err = algs.ReadClass(tx, cubeID, classID)
			return err
		}, &sql.TxOptions{ReadOnly: true})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		resp := AlgorithmGroupsWithCubeResponse{AlgorithmClass: *class}
		svc.Cache.Set(key, resp, 120*time.Minute)
		ctx.JSON(http.StatusOK, resp)
	}
}
