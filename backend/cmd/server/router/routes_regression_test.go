package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type routeKey struct {
	method string
	path   string
}

func collectRoutes() []gin.RouteInfo {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api/v1")
	setupAllRoutes(apiGroup)
	return r.Routes()
}

func hasRoute(routes []gin.RouteInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func TestSetupAllRoutes_KeyRoutesPresent(t *testing.T) {
	routes := collectRoutes()

	cases := []routeKey{
		{method: "GET", path: "/api/v1/alert/alarms"},
		{method: "POST", path: "/api/v1/alert/triggers"},
		{method: "GET", path: "/api/v1/monitoring/monitors"},
		{method: "POST", path: "/api/v1/monitoring/items/history-generations"},
		{method: "GET", path: "/api/v1/delivery/media"},
		{method: "POST", path: "/api/v1/delivery/media/qq/messages"},
		{method: "GET", path: "/api/v1/system/logs/system"},
		{method: "DELETE", path: "/api/v1/system/logs/:type"},
		{method: "GET", path: "/api/v1/users"},
		{method: "POST", path: "/api/v1/users/sessions"},
		{method: "GET", path: "/api/v1/analysis/reports"},
		{method: "POST", path: "/api/v1/analysis/alert-storms"},
		{method: "GET", path: "/api/v1/ai/chats"},
		{method: "POST", path: "/api/v1/ai/providers/:id/checks"},
		{method: "POST", path: "/api/v1/ai/mcp/messages"},
	}

	for _, tc := range cases {
		if !hasRoute(routes, tc.method, tc.path) {
			t.Fatalf("missing route: %s %s", tc.method, tc.path)
		}
	}
}

func TestSetupAllRoutes_NoDuplicateMethodPath(t *testing.T) {
	routes := collectRoutes()
	seen := make(map[routeKey]struct{}, len(routes))

	for _, route := range routes {
		key := routeKey{method: route.Method, path: route.Path}
		if _, ok := seen[key]; ok {
			t.Fatalf("duplicate route detected: %s %s", route.Method, route.Path)
		}
		seen[key] = struct{}{}
	}

	if len(routes) < 50 {
		t.Fatalf("unexpectedly low number of routes: got %d", len(routes))
	}

	if len(seen) != len(routes) {
		t.Fatalf("route key mismatch: seen=%d routes=%d", len(seen), len(routes))
	}
}
