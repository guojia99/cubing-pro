package groupcomp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Parse route declarations without importing svc, whose transitive init starts remote jobs.
func TestAdminRouteContract(t *testing.T) {
	path := filepath.Join("..", "..", "api", "routes", "admin.go")
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		recv, ok := sel.X.(*ast.Ident)
		if !ok || recv.Name != "managed" || len(call.Args) == 0 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if ok {
			found[sel.Sel.Name+" "+strings.Trim(lit.Value, `"`)] = true
		}
		return true
	})
	for _, route := range []string{"GET ", "POST ", "DELETE ", "POST /apply", "POST /end", "GET /players", "POST /players/approval/:reg_id", "GET /result", "POST /result", "DELETE /result/:result_id", "GET /pre_results", "POST /pre_results/:result_id/approval", "PUT /refresh_event"} {
		if !found[route] {
			t.Errorf("missing route %s", route)
		}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `admin.Group("/competition")`) || !strings.Contains(string(content), `Group("/comps/:compId", org_mid.AdminCompMiddleware(svc))`) {
		t.Fatal("admin group must resolve actual competition owner")
	}
	if !strings.Contains(string(content), "middleware.CheckAuthMiddlewareFunc(user2.AuthAdmin, user2.AuthSuperAdmin)") {
		t.Fatal("both admin roles required")
	}
	if !strings.Contains(string(content), `compWithAdmin.GET("/filters", organizers.CompetitionFilters(svc))`) {
		t.Fatal("filter options must be on admin group rather than super-only group")
	}
	if !strings.Contains(string(content), `managed.POST("/pre_results/:result_id/approval", organizers.AdminApprovalCompPlayerPreResult(svc))`) {
		t.Fatal("admin approval capability must use dedicated admin handler")
	}
}

func TestOrganizerCannotUseHistoricalConfirmation(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "api", "routes", "comp_result.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "AdminApprovalCompPlayerPreResult") {
		t.Fatal("admin handler exposed on organizer route")
	}
	content, err = os.ReadFile(filepath.Join("..", "..", "api", "app", "organizers", "ApprovalCompPlayerPreResult.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "return approvalCompPlayerPreResult(svc, false)") || !strings.Contains(string(content), "req.HistoricalConfirmed && !adminRoute") {
		t.Fatal("organizer route must reject historical confirmation even for an admin user")
	}
}

func TestPlayerRouteContract(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "api", "routes", "comp_result.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	start := strings.Index(s, `"/player_comp",`)
	if start < 0 {
		t.Fatal("player group missing")
	}
	end := strings.Index(s[start:], "results := userComp.Group")
	if end < 0 {
		t.Fatal("group declaration incomplete")
	}
	block := s[start : start+end]
	jwt, role := strings.Index(block, "middleware.JWT().MiddlewareFunc()"), strings.Index(block, "middleware.CheckAuthMiddlewareFunc")
	if jwt < 0 || role < 0 || jwt > role {
		t.Fatal("JWT must precede player role authorization")
	}
	if strings.Contains(s, `/:reg_id/:compId/refresh_event`) || !strings.Contains(s, `PUT("/refresh_event"`) {
		t.Fatal("refresh_event route malformed")
	}
}
