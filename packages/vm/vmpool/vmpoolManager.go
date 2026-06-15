package vmpool

import (
	"context"
	"sync"
	"time"

	"github.com/omjikush09/sandboxing-infra/packages/vm/start"
)

var pool = VmPool{
	mu:     sync.Mutex{},
	idleVM: make(map[string]*start.VM),
	inUse:  make(map[string]*start.VM),
}

func InitPoolManager(ctx context.Context, idleSize int) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pool.Clear()
			return
		case <-ticker.C:
			idealVmSize := pool.GetSizeIdelVM()

			for i := idealVmSize; i < idleSize; i++ {
				pool.AddToPool()
			}
		}
	}
}

func GetInstance() (*start.VM, bool) {
	return pool.GetaVM()
}

func DeleteVM(vm *start.VM) {
	pool.RemoveFromPool(vm)
}
