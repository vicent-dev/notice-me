package app

import (
	"testing"

	"github.com/gorilla/mux"
)

func TestRoutesRegistersExpectedEndpoints(t *testing.T) {
	initialiseMocks()
	s.r = mux.NewRouter()

	s.routes()

	expected := map[string]bool{
		"GET /ws":                            false,
		"GET /api/docs":                      false,
		"POST /api/notifications":            false,
		"GET /api/notifications":             false,
		"GET /api/notifications/notify/{id}": false,
		"GET /api/notifications/{id}":        false,
		"DELETE /api/notifications/{id}":     false,
	}

	err := s.r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		methods, _ := route.GetMethods()
		for _, m := range methods {
			key := m + " " + path
			if _, ok := expected[key]; ok {
				expected[key] = true
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("walking routes: %v", err)
	}

	for key, found := range expected {
		if !found {
			t.Errorf("route not registered: %s", key)
		}
	}
}
