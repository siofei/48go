package LiveRoom

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Pool struct {
	Mutex sync.Mutex
	wait  sync.WaitGroup
	max   int
	P     map[int]*Live
	C     chan int
}

func New(max int) *Pool {
	var pool = Pool{}
	pool.max = max
	pool.C = make(chan int, max)
	pool.wait = sync.WaitGroup{}
	return &pool
}

type Exist error
type Pull error

func (p *Pool) ShowDownloading(ch chan interface{}) {
	for {
		time.Sleep(10 * time.Minute)
		if len(p.P) == 0 {
			continue
		}
		ch <- fmt.Sprintf("\n当前正在下载:%d", len(p.P))
		for _, v := range p.P {
			ch <- *v
		}
		ch <- "\n"
	}
}

func (p *Pool) Add(l *Live) error {
	if _, ok := p.P[l.LiveId.Int()]; ok {
		return Exist(fmt.Errorf("%s is exist", l.LiveId))
	}
	if len(p.P) >= p.max {
		return Pull(errors.New("pull"))
	}
	if p.P == nil {
		p.P = make(map[int]*Live, p.max)
	}
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.C <- 1
	p.P[l.LiveId.Int()] = l
	p.wait.Add(1)
	return nil
}

func (p *Pool) Done(id int) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	delete(p.P, id)
	<-p.C
	p.wait.Done()
}

func (p *Pool) Wait() {
	p.wait.Wait()
}
