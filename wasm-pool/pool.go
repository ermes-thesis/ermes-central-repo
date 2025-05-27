package wasm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
	"os"
	"sync"
)

type WasmExecutor struct {
	vm    *wasmedge.VM
	mu    sync.Mutex
	inUse bool
}

type ExecutorPool struct {
	instances []*WasmExecutor
	mu        sync.Mutex
}

func NewExecutorPool(poolSize int) *ExecutorPool {
	pool := &ExecutorPool{}
	for i := 0; i < poolSize; i++ {
		wasmedge.SetLogErrorLevel()
		conf := wasmedge.NewConfigure(wasmedge.WASI)
		vm := wasmedge.NewVMWithConfig(conf)
		wasi := vm.GetImportModule(wasmedge.WASI)
		wasi.InitWasi(
			os.Args[1:],
			os.Environ(),
			[]string{".:."},
		)

		pool.instances = append(pool.instances, &WasmExecutor{vm: vm, inUse: false})
	}
	return pool
}

func (p *ExecutorPool) Acquire() *WasmExecutor {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, executor := range p.instances {
		if !executor.inUse {
			executor.inUse = true
			return executor
		}
	}

	return nil
}

func (p *ExecutorPool) Release(executor *WasmExecutor) {
	executor.mu.Lock()
	defer executor.mu.Unlock()

	executor.inUse = false
}
