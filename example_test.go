package atomiccounter_test

import (
	"fmt"
	"sync"

	"github.com/chen3feng/atomiccounter"
)

func Example() {
	counter := atomiccounter.NewInt64()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			counter.Inc()
			wg.Done()
		}()

	}
	wg.Wait()
	fmt.Println(counter.Load())
	counter.Set(0)
	fmt.Println(counter.Load())
	// Output:
	// 100
	// 0
}
