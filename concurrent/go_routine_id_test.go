package concurrent

import (
	"fmt"
	"sync"
	"testing"
)

func TestGoroutineID(t *testing.T) {
	fmt.Printf("Start main routine %v \n", GoroutineID())
	max := 2200 // FIXME - upper limit, more than 2200 routine the method GoroutineID throw a panic error
	allGoRoutineStart := sync.WaitGroup{}
	allGoRoutineStart.Add(max)
	allGoRoutineEnd := sync.WaitGroup{}
	allGoRoutineEnd.Add(max)
	for i := range max {
		go func() {
			fmt.Printf("Start go routine %v -> %v \n", i, GoroutineID())
			defer allGoRoutineEnd.Done()
			for j := range 10000 {
				if j == 0 {
					allGoRoutineStart.Done()
				}
				GoroutineID()
			}
		}()
	}
	allGoRoutineStart.Wait()
	fmt.Printf("Waiting for %v go routines ...\n", max)
	allGoRoutineEnd.Wait()

}
