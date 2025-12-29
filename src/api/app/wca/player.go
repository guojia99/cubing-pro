package wca

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func Players(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("name")
		out := svc.Wca.SearchPlayers(name)
		ctx.JSON(http.StatusOK, out)
	}
}

func PlayerResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, _ := svc.Wca.GetPersonResult(p)

		sort.Slice(out, func(i, j int) bool {
			if out[i].CompetitionID == out[j].CompetitionID {
				return out[i].Pos < out[j].Pos
			}
			return out[i].ID > out[j].ID
		})
		ctx.JSON(http.StatusOK, out)
	}
}

func PlayerCompetitions(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, _ := svc.Wca.GetPersonCompetition(p)
		ctx.JSON(http.StatusOK, out)
	}
}

func PlayerPersonInfo(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Param("wcaID")

		out, err := svc.Wca.GetPersonInfo(p)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{})
			return
		}
		ctx.JSON(http.StatusOK, out)
	}
}
