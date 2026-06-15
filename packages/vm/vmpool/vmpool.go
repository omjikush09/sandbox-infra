package vmpool

import (
	"log/slog"
	"sync"

	"github.com/omjikush09/sandboxing-infra/packages/vm/start"
)

type VmPool struct {
	mu     sync.Mutex
	idleVM map[string]*start.VM
	inUse  map[string]*start.VM
}

func (p *VmPool) GetSizeIdelVM() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.idleVM)

}

func (p *VmPool) AddToPool() {

	vm, err := start.CreateVm()

	if err != nil {
		slog.Error("Failed to create VM ", "err", err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.idleVM[vm.Id] = vm

}

func (p *VmPool) GetaVM() (*start.VM, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k, v := range p.idleVM {
		p.markAsInUse(k)
		return v, true
	}
	return nil, false
}

func (p *VmPool) markAsInUse(id string) {

	vm, ok := p.idleVM[id]
	if !ok {
		return
	}
	p.inUse[id] = vm
	delete(p.idleVM, id)
}

func (p *VmPool) RemoveFromPool(vm *start.VM) {
	if vm == nil {
		return
	}
	p.destroyVM(vm)

	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.inUse, vm.Id)
	delete(p.idleVM, vm.Id)

}

func (p *VmPool) destroyVM(vm *start.VM) {
	vm.Cleanup()
}

func (p *VmPool) Clear() {

	inUse := []*start.VM{}
	idleVM := []*start.VM{}
	p.mu.Lock()
	for _, v := range p.inUse {
		inUse = append(inUse, v)
	}
	clear(p.inUse)
	for _, v := range p.idleVM {
		idleVM = append(idleVM, v)
	}
	clear(p.idleVM)
	p.mu.Unlock()
	for _, v := range inUse {
		p.destroyVM(v)
	}
	for _, v := range idleVM {
		p.destroyVM(v)
	}
}
