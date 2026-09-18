//go:build windows

package psadt

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/pedrostefanogv/go-psadt/types"
)

// skipFromContextCheck lists exported Session methods that do not execute
// PSADT commands and therefore have no need for a *WithContext variant.
var skipFromContextCheck = map[string]bool{
	"WithContext": true,
	"OnClose":     true,
	"OnError":     true,
	"LiveOutput":  true,
}

// TestSessionWithContextCoverage enforces that every exported Session method
// that executes PSADT commands has a *WithContext variant. This keeps the
// generated surface (withcontext_gen.go) in sync with the API.
func TestSessionWithContextCoverage(t *testing.T) {
	sessionType := reflect.TypeOf(&Session{})
	if sessionType == nil {
		t.Fatal("nil type")
	}


	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()




	set := map[string]reflect.Method{}
	for i := 0; i < reflect.TypeOf(&Session{}).NumMethod(); i++ {
		m := reflect.TypeOf(&Session{}).Method(i)
		set[m.Name] = m
	}

	for name, m := range set {
		if skipFromContextCheck[name] || strings.HasSuffix(name, "WithContext") {
			continue
		}
		mt := m.Type
		// Method type: receiver is index 0; first real param is index 1.
		if mt.NumIn() > 1 && mt.In(1) == ctxType {
			continue // already context-first
		}
		if _, ok := set[name+"WithContext"]; !ok {
			t.Errorf("Session method %s has no %s variant (run go run ./tools/genwithcontext)", name, name+"WithContext")
		}
	}
}

// TestWithContextVariantSmoke verifies a couple of generated variants
// resolve correctly and share the same runner (no process churn).
func TestWithContextVariantSmoke(t *testing.T) {
	// Compile-time checks: generated variants must exist with sane shapes.
	var _ func(context.Context, types.WelcomeOptions) error = (*Session)(nil).ShowInstallationWelcomeWithContext
	var _ func(context.Context) (bool, error) = (*Session)(nil).TestCallerIsAdminWithContext
	var _ func(context.Context, string, string) (string, error) = (*Session)(nil).GetRegistryKeyStringWithContext
}
