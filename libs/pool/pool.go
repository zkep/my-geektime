package pool

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

type WorkUnit struct {
	ctx   context.Context
	once  sync.Once
	ch    chan *any
	work  WorkFunc
	Value any
	Err   error
}

func (wu *WorkUnit) Get() (any, error) {
	select {
	case <-wu.ch:
		return wu.Value, wu.Err
	case <-wu.ctx.Done():
		return wu.Value, wu.Err
	}
}

func (wu *WorkUnit) AttachValue(val any) {
	wu.once.Do(func() {
		wu.Value = val
		close(wu.ch)
	})
}

func (wu *WorkUnit) Error(err error) {
	wu.once.Do(func() {
		wu.Err = err
		close(wu.ch)
	})
}

type WorkFunc func(pctx context.Context) (any, error)

type GPool struct {
	limiter chan any
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewLimitPool(ctx context.Context, maxcapacity int) *GPool {
	limiter := make(chan any, maxcapacity)
	ctx, cancel := context.WithCancel(ctx)
	return &GPool{
		ctx:     ctx,
		cancel:  cancel,
		limiter: limiter,
	}
}

var (
	ErrQueueTimeout     = errors.New("queue timeout")
	ErrQueueContextDone = errors.New("context is done")
)

func (p *GPool) Queue(ctx context.Context, work WorkFunc) (*WorkUnit, error) {
	wu := &WorkUnit{ch: make(chan *any, 1), work: work, ctx: ctx}
	return wu, p.queue(wu)
}

func (p *GPool) queue(wu *WorkUnit) error {
	select {
	case <-wu.ctx.Done():
		wu.Error(ErrQueueContextDone)
		return wu.Err
	case p.limiter <- nil:
		go func() {
			defer func() {
				<-p.limiter
				if err := recover(); nil != err {
					fmt.Println("handler", "GPool|Queue|Panic|%v|%s", err, string(debug.Stack()))
					wu.Error(fmt.Errorf("%v", err))
				}
			}()

			select {
			case <-p.ctx.Done():
				wu.Error(ErrQueueContextDone)
				return
			case <-wu.ctx.Done():
				wu.Error(ErrQueueContextDone)
				return
			default:
			}
			val, err := wu.work(wu.ctx)
			if nil != err {
				wu.Error(err)
			} else {
				wu.AttachValue(val)
			}
		}()
	case <-p.ctx.Done():
		wu.Error(ErrQueueContextDone)
		return wu.Err
	}
	return nil
}

func (p *GPool) Monitor() (int, int) {
	return len(p.limiter), cap(p.limiter)
}

func (p *GPool) Close() {
	p.cancel()
	close(p.limiter)
}

type Batch struct {
	gopool *GPool
	works  []WorkFunc
}

func (p *GPool) NewBatch() *Batch {
	return &Batch{
		gopool: p,
		works:  make([]WorkFunc, 0, 5),
	}
}

func (p *Batch) Queue(work WorkFunc) *Batch {
	p.works = append(p.works, work)
	return p
}

func (p *Batch) Wait(ctx context.Context) ([]*WorkUnit, error) {
	wus := make([]*WorkUnit, 0, len(p.works))
	for i := range p.works {
		wu := &WorkUnit{
			ctx:  ctx,
			work: p.works[i],
			ch:   make(chan *any, 1),
		}
		_ = p.gopool.queue(wu)

		wus = append(wus, wu)
	}

	for i := range wus {
		select {
		case <-p.gopool.ctx.Done():
			for j := i; j < len(wus); j++ {
				if nil == wus[j].Err {
					wus[j].Error(ErrQueueTimeout)
				}
			}
			return wus, nil
		case <-wus[i].ctx.Done():
			for j := i; j < len(wus); j++ {
				if nil == wus[j].Err {
					wus[j].Error(ErrQueueTimeout)
				}
			}
			return wus, nil
		case <-wus[i].ch:
		}
	}
	return wus, nil
}
