//go:build windows

package main

import "testing"

func TestRecordClientInvalidationStoresReason(t *testing.T) {
	app := &appState{}

	app.recordClientInvalidation("runner caiu")

	if app.page.ClientReady {
		t.Fatal("expected client to be marked as not ready")
	}
	if app.page.ClientStatus != "error" {
		t.Fatalf("expected status error, got %q", app.page.ClientStatus)
	}
	if app.page.LastClientInvalidationReason != "runner caiu" {
		t.Fatalf("expected invalidation reason to be stored, got %q", app.page.LastClientInvalidationReason)
	}
	if app.page.LastClientInvalidationAt == "" {
		t.Fatal("expected invalidation timestamp to be recorded")
	}
}

func TestRecordClientConnectedCountsOnlyReconnects(t *testing.T) {
	app := &appState{}

	app.recordClientConnected(false)
	if !app.page.ClientReady {
		t.Fatal("expected client to be ready after first connect")
	}
	if app.page.ClientStatus != "ready" {
		t.Fatalf("expected status ready, got %q", app.page.ClientStatus)
	}
	if app.page.ClientReconnects != 0 {
		t.Fatalf("expected reconnect count 0 after initial connect, got %d", app.page.ClientReconnects)
	}

	app.recordClientInvalidation("runner caiu")
	app.recordClientConnected(true)
	if app.page.ClientReconnects != 1 {
		t.Fatalf("expected reconnect count 1 after reconnect, got %d", app.page.ClientReconnects)
	}
	if app.page.LastClientInvalidationReason != "runner caiu" {
		t.Fatalf("expected last invalidation reason to remain available, got %q", app.page.LastClientInvalidationReason)
	}
}
