//go:build windows

package psadt

import (
	"context"
	"fmt"
	"sync"
)

// ClientPool maintains a pool of ready-to-use PSADT clients for parallel
// deployments. Each Client owns its own persistent PowerShell process, so
// parallel deployments require one client each — the pool amortizes the
// process start + Import-Module cost across tasks.
//
// Example:
//
//	pool, err := psadt.NewClientPool(ctx, 4, psadt.WithAutoReconnect())
//	defer pool.Close()
//	client, _ := pool.Acquire(ctx)
//	defer pool.Release(client)
type ClientPool struct {
	opts []Option

	mu      sync.Mutex
	idle    []*Client
	all     []*Client
	maxSize int
	closed  bool
}

// NewClientPool creates a pool pre-warmed with size clients. When Acquire
// exhausts the idle clients, new ones are created on demand (bounded only by
// memory — call Acquire with a ctx deadline to bound the start latency).
func NewClientPool(ctx context.Context, size int, opts ...Option) (*ClientPool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("pool size must be > 0")
	}
	p := &ClientPool{opts: opts, maxSize: size * 4}
	for i := 0; i < size; i++ {
		c, err := NewClientWithContext(ctx, opts...)
		if err != nil {
			// Roll back the clients already started.
			p.Close()
			return nil, fmt.Errorf("failed to start pool client %d/%d: %w", i+1, size, err)
		}
		p.all = append(p.all, c)
		p.idle = append(p.idle, c)
	}
	return p, nil
}

// Acquire returns an idle, alive client — starting a new one when the pool
// is exhausted. The returned client must be given back with Release.
func (p *ClientPool) Acquire(ctx context.Context) (*Client, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("client pool is closed")
	}
	if n := len(p.idle); n > 0 {
		c := p.idle[n-1]
		p.idle = p.idle[:n-1]
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()

	c, err := NewClientWithContext(ctx, p.opts...)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.all = append(p.all, c)
	p.mu.Unlock()
	return c, nil
}

// Release returns a client to the pool. Dead clients are replaced
// transparently: the next Acquire starts a fresh one if needed.
func (p *ClientPool) Release(c *Client) {
	if c == nil {
		return
	}
	if !c.IsAlive() {
		// Best effort recovery; on failure just close it. Acquire will
		// create a replacement when the idle list runs dry.
		ctx, cancel := context.WithTimeout(context.Background(), defaultInitTimeout)
		defer cancel()
		if err := c.Reconnect(ctx); err != nil {
			c.Close()
			return
		}
	}
	p.mu.Lock()
	if !p.closed && len(p.all) <= p.maxSize {
		p.idle = append(p.idle, c)
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()
	c.Close()
}

// Close shuts down every client the pool ever created.
func (p *ClientPool) Close() {
	p.mu.Lock()
	p.closed = true
	all := make([]*Client, len(p.all))
	copy(all, p.all)
	p.all = nil
	p.idle = nil
	p.mu.Unlock()

	for _, c := range all {
		c.Close()
	}
}

// Size returns the current number of live clients managed by the pool.
func (p *ClientPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.all)
}
