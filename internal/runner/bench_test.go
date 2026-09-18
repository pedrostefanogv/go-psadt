//go:build windows

package runner

import (
	"context"
	"testing"
	"time"
)

// BenchmarkRoundTrip measures the per-command latency through the persistent
// PowerShell process (the amortized cost that justifies the design).
func BenchmarkRoundTrip(b *testing.B) {
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		b.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	ctx := context.Background()
	// Warm-up
	if _, err := r.Execute(ctx, "$true"); err != nil {
		b.Fatalf("warm-up failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := r.Execute(ctx, "$true"); err != nil {
			b.Fatalf("round-trip failed: %v", err)
		}
	}
}

// BenchmarkRoundTripVoid measures void-command latency (no data parsing).
func BenchmarkRoundTripVoid(b *testing.B) {
	r, err := New(Config{Timeout: 30 * time.Second})
	if err != nil {
		b.Fatalf("failed to start runner: %v", err)
	}
	defer r.Stop()

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := r.ExecuteVoid(ctx, "$null = $null"); err != nil {
			b.Fatalf("void round-trip failed: %v", err)
		}
	}
}
