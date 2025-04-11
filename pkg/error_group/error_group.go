package errorgroup

import (
	"context"
	"fmt"
	"sync"
)

type tokenLimit struct{}

type ProductGroup struct {
	cancel func(error)
	wg     sync.WaitGroup
	limit  chan tokenLimit
	err    error
}

// done разблокирует горутину из лимита
func (g *ProductGroup) done() {
	if g.limit != nil {
		<-g.limit
	}
	g.wg.Done()
}

// Создаёт группу с функцией записи ошибки в контекст
//
// Группа содержит полезную нагрузку в виде payload *model.Good
func EGWithContext(ctx context.Context) (*ProductGroup, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &ProductGroup{cancel: cancel}, ctx
}

// Wait блокирует, пока не вернутся все вызовы из Go,
// затем возвращает первую не ниловую ошибку, если она вернулась в одном из вызовов
// также возвращает полезную нагрузку из вызывающей функции, если она была
func (g *ProductGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}

	return g.err
}

// Go вызывает переданную функцию в новой горутине
// ProductGroup не блокируется до момента, пока не будет превышен заданный лимит для группы
// error и payload будут возвращены методом Wait
func (g *ProductGroup) Go(f func() error) {
	if g.limit != nil {
		g.limit <- tokenLimit{}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		g.err = f()
	}()
}

// SetLimit устанавливает кол-во активных горутин в группе
// Отрицательное значение убирает лимит
// Вызов Go метода будет заблокирован, пока он сможет добавить горутину в рамках лимита
func (g *ProductGroup) SetLimit(n int) {
	if n < 0 {
		g.limit = nil
		return
	}
	if len(g.limit) != 0 {
		panic(fmt.Errorf("errgroup: try to modify limit while %v goroutines are still active", len(g.limit)))
	}
	g.limit = make(chan tokenLimit, n)
}
