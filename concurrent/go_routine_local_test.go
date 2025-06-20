package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func TestGoRoutineLocal(t *testing.T) {
	max := 1400 // Upper limit, because a limit en GoRoutineId() method.
	grl := GoRoutineLocal[int]{}
	wg := sync.WaitGroup{}
	wg.Add(max)
	for i := range max {
		local := i
		if i%2 == 0 {
			go func() {
				grl.Set(&local)
				fmt.Printf("Start go routine %v\n", local)
				defer wg.Done()
				for range 10000 {
					stored := grl.Get()
					if stored == nil || *stored != local {
						t.Fail()
						return
					}
				}
			}()
		} else {
			go func() {
				fmt.Printf("Start go routine %v\n", local)
				defer wg.Done()
				for range 10000 {
					stored := grl.Get()
					if stored != nil {
						t.Fail()
						return
					}
				}
			}()

		}
	}
	fmt.Printf("Waiting for %v go routines ...\n", max)
	wg.Wait()

}
