package lock

import (
	"context"
	"sync"
	"sync/atomic"
)

// lock для получения доступа к ИИ и использование ресурсов Cpu/Mem

var Resource = newResourceLock()

func InitResourceLock(ctx context.Context){
	Resource = newResourceLock()

	go func() {
		<-ctx.Done()
		Resource.Stop()
	}()
}

/*
В AI функциях
func AIFunction(ctx context.Context) {
	if !Resource.Acquire(ctx, "AIFunction") {
		return // Контекст завершен
	}
	defer Resource.Release("AIFunction")

	// Работа с ИИ...
}
Так, только одна AI функция будет выполняться в любой момент времени, а остальные будут ждать своей очереди или завершаться при отмене контекста 
*/

type ResourceLock struct {
	mu         sync.Mutex
	cond       *sync.Cond
	holder     string
	waitCount  int32
	stopCh     chan struct{}
	stopped    bool
}

func newResourceLock() *ResourceLock {
	lock := &ResourceLock{
		stopCh: make(chan struct{}),
	}
	lock.cond = sync.NewCond(&lock.mu)
	return lock
}

// Acquire пытается захватить ресурс для указанной функции
// Возвращает true если ресурс получен, false если контекст завершился
func (c *ResourceLock) Acquire(ctx context.Context, functionName string) bool {
	atomic.AddInt32(&c.waitCount, 1)
	defer atomic.AddInt32(&c.waitCount, -1)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopped {
		return false
	}

	for c.holder != "" && !c.stopped {
		select {
		case <-ctx.Done():
			return false
		default:
			// Используем Wait с таймаутом для периодической проверки контекста
			c.cond.Wait()
		}
	}

	if c.stopped {
		return false
	}

	c.holder = functionName
	return true
}

// Release освобождает ресурс
func (c *ResourceLock) Release(functionName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.holder == functionName {
		c.holder = ""
		c.cond.Broadcast()
	}
}

// Stop останавливает все ожидающие горутины
func (c *ResourceLock) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopped = true
	close(c.stopCh)
	c.cond.Broadcast()
}

// WaitCount возвращает количество ожидающих горутин
func (c *ResourceLock) WaitCount() int {
	return int(atomic.LoadInt32(&c.waitCount))
}
