package baseworker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"runtime/debug"
	"time"
)

type BaseImpl struct {
	WorkerName    string
	firstRunDelay time.Duration
	runInterval   time.Duration
}

func NewInstance(WorkerName string, firstRunDelay, runInterval time.Duration) *BaseImpl {
	return &BaseImpl{
		WorkerName:    WorkerName,
		firstRunDelay: firstRunDelay,
		runInterval:   runInterval,
	}
}

func (i BaseImpl) GetLogger() *log.Entry {
	logger := log.
		WithField("worker_name", i.WorkerName)
	return logger
}

func (i BaseImpl) Run(ctx context.Context, jobFunc func(ctx context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			i.GetLogger().
				WithField("panic_stack", string(debug.Stack())).
				Errorf("panic: (%v)", r)
		}
	}()
	period := i.firstRunDelay
	logger := i.GetLogger()
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			logger.Info("Задача остановлена")
			return
		case <-time.After(period):
			logger.Info("Задача запущена")
			jobFunc(ctx)
			logger.Info("Задача выполнена")
		}
		period = i.runInterval
	}
}
