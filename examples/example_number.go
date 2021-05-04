package examples

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/keithwachira/go-taskq"
)

// AddNumbers sample of how you can sum numbers
func AddNumbers() (int, int) {
	done := make(chan bool)
	complete:=0
	initialValue := 100
	sum := 0
	var mutex = &sync.Mutex{}
	q := taskq.NewQueue(1, 4, func(job interface{}) {
		if number, ok := job.(int); ok {
			mutex.Lock()
			sum += number
			initialValue -= number
			complete+=1
			if complete == 9 {
				//finished processing numbers
				done <- true
			}
			mutex.Unlock()
		}
	})

	go q.StartWorkers()
	for i := 0; i < 10; i++ {
		q.EnqueueJobBlocking(i)
	}
	<-done
	return sum, initialValue

}

func ShowBusyWait() {
	index := 0
	done := make(chan bool)
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	queue := taskq.NewQueue(1, 5, func(job interface{}) {
		time.Sleep(10 * time.Second)

		if index == len(items) {
			done <- true
		}
		fmt.Println("i have finished processing item item:", job)
	})
	go queue.StartWorkers()

out:
	for {

		for index < len(items) {
			item := items[index]
			queued := queue.EnqueueJobNonBlocking(item)
			if queued {
				log.Println("item added was", item)
				index += 1
				if index == len(items) {

					//finished queueing all items break from outer loop
					break out
				}
			} else {
				log.Println("item was not queued unnecessary loop here")

			}

		}

	}
	//to wait for workers to finish
	<-done

	//time.Sleep(50 * time.Second)
}
