package infrastructurejsengine

import (
	"fmt"
	"time"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	"github.com/dop251/goja"
)

type gojaImpl struct{}

func NewGojaImpl() domaincontractsutility.JSEngine {
	return &gojaImpl{}
}

func (g *gojaImpl) RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error) {
	vm := goja.New()

	timer := time.AfterFunc(timeout, func() {
		vm.Interrupt(fmt.Errorf("encoder execution exceeded %s", timeout))
	})
	defer timer.Stop()

	if _, err := vm.RunString(source); err != nil {
		return nil, fmt.Errorf("failed to load encoder source: %w", err)
	}

	encodeFn, ok := goja.AssertFunction(vm.Get("encode"))
	if !ok {
		return nil, fmt.Errorf("encoder source does not define a callable encode function")
	}

	stateValue := vm.ToValue(state)
	result, err := encodeFn(goja.Undefined(), stateValue)
	if err != nil {
		return nil, fmt.Errorf("encoder threw an error: %w", err)
	}

	exported := result.Export()
	items, ok := exported.([]interface{})
	if !ok {
		return nil, fmt.Errorf("encoder returned %T, want an array of numbers", exported)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("encoder returned an empty array")
	}

	durations := make([]int32, len(items))
	for i, item := range items {
		switch v := item.(type) {
		case int64:
			durations[i] = int32(v)
		case float64:
			durations[i] = int32(v)
		default:
			return nil, fmt.Errorf("encoder array element %d is %T, want a number", i, item)
		}
	}
	return durations, nil
}
