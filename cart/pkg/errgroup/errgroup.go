package errgroup

import (
	"context"
	"sync"
)

type Group struct {
	wg     sync.WaitGroup
	cancel func()
	err    error
	mu     sync.Mutex
	sem    chan struct{}
}

func WithContext(ctx context.Context, limit int) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	g := &Group{cancel: cancel, sem: make(chan struct{}, limit)}

	return g, ctx
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	close(g.sem)
	return g.err
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		g.sem <- struct{}{}
		defer func() {
			<-g.sem
			g.wg.Done()
		}()

		if err := f(); err != nil {
			g.mu.Lock()
			if g.err == nil {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			}
			g.mu.Unlock()
		}
	}()
}
