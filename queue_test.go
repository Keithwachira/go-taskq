package taskq

import (
	"sync"
	"testing"
)

func TestJobQueueCapacity(t *testing.T) {
	q := NewQueue(1, 4, func(job interface{}) {})

	checkIntEqual(t, "Queue capacity ", 4, cap(q.JobQueue))
	q50 := NewQueue(1, 50, func(job interface{}) {})

	checkIntEqual(t, "Queue capacity ", 50, cap(q50.JobQueue))

}
func TestQueue_EnqueueJobNonBlocking(t *testing.T) {
	rejectedJobs := []int{}
	q := NewQueue(1, 4, func(job interface{}) {

	})
	for i := 0; i < 10; i++ {
		queued := q.EnqueueJobNonBlocking(i)
		if !queued {
			rejectedJobs = append(rejectedJobs, 1)
		}

	}

	checkIntEqual(t, "Length JobQueue queued jobs ", 4, len(q.JobQueue))
	checkIntEqual(t, "Length JobQueue  rejected jobs ", 6, len(rejectedJobs))

}
func checkIntEqual(t *testing.T, name string, expected, actual int) {
	if actual != expected {
		t.Errorf("%s returned unexpected value: got %d wanted %d",
			name, actual, expected)
	}
}

func TestQueue_EnqueueJobBlocking(t *testing.T) {
	queuedJobs := []int{}
	rejectedJobs := []int{}
	done := make(chan bool)

	q := NewQueue(1, 4, func(job interface{}) {
		if number, ok := job.(int); ok {
			if number == 9 {
				//finished processing numbers

				done <- true
			}
		}
	})

	go q.StartWorkers()
	for i := 0; i < 10; i++ {
		q.EnqueueJobBlocking(i)
		queuedJobs = append(queuedJobs, i)
	}
	<-done

	checkIntEqual(t, "Length JobQueue queued jobs ", 10, len(queuedJobs))
	checkIntEqual(t, "Length JobQueue  rejected jobs ", 0, len(rejectedJobs))

}

func TestQueue_JobCallBack(t *testing.T) {
	//we will subtract sum of all numbers from 100
	//to test if our jobcallback is fine
	done := make(chan bool)
	complete := 0
	initialValue := 100
	sum := 0
	var mutex = &sync.Mutex{}
	q := NewQueue(4, 4, func(job interface{}) {
		if number, ok := job.(int); ok {
			mutex.Lock()
			sum += number
			initialValue -= number
			complete += 1
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

	checkIntEqual(t, "Sum 0 to n ", 45, sum)
	checkIntEqual(t, "100 minus sum ", 55, initialValue)

}
