package clickhouse

import (
	"sync/atomic"
	"fmt"
	composer "github.com/leprosus/golang-composer"
	"sync"
)

type Limiter struct {
	once sync.Once

	maxRequests     uint32
	requestsCounter int32
	queue           chan int32
}

// Sets requests limitation (zero is limitation off)
func (lim *Limiter) MaxRequests(limit int) {
	atomic.StoreUint32(&lim.maxRequests, uint32(limit))

	message := fmt.Sprintf("Set max request pool = %d", limit)
	cfg.logger.debug(message)

	lim.once.Do(func() {
		lim.queue = make(chan int32)

		go func() {
			for step := range lim.queue {
				atomic.AddInt32(&lim.requestsCounter, step)

				if atomic.LoadUint32(&lim.maxRequests) > 0 {
					if atomic.LoadUint32(&lim.maxRequests) <= uint32(atomic.LoadInt32(&lim.requestsCounter)) {
						composer.GetComposer().Pause()
					} else {
						composer.GetComposer().Play()
					}
				} else {
					composer.GetComposer().Play()
				}
			}
		}()
	})
}

func (lim *Limiter) increase() {
	lim.queue <- 1
}

func (lim *Limiter) reduce() {
	lim.queue <- -1
}

func (lim *Limiter) waitForRest() {
	composer.GetComposer().NeedWait()
}
