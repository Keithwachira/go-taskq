package taskq

import (
	"log"
	"sync"
)

// JobCallBack work on the queued item
/*type JobCallBack interface {
	Process(interface{})
}*/


type JobCallBack func(job interface{})
type Queue struct {
	//Workers Number of goroutines(workers,consumers) to be used to process the jobs
	Workers int
	//Capacity is the number of  items that will be held in the JobQueue channel (its capacity) at a time before blocking
	//i.e capacity of our JobQueue channel
	Capacity int
	//JobQueue a buffered channel of capacity [Queue.Capacity] that will temporary hold
	//jobs before they are assigned to a worker
	JobQueue chan interface{}
	//Wg will be used to make sure the program does not terminate before all of our goroutines
	//complete the job assigned to them
	Wg *sync.WaitGroup
	///timeout in seconds
	///TimeOut  time.Duration
	//QuitChan will be used to stop all goroutines [Queue.Workers]
	QuitChan chan struct{}
	//JobCallBack is the function to be called when a job is received
	//it should implement JobCallBack
	//when a job is received what should happen ?(call back)
	JobCallBack JobCallBack
}

// NewQueue create a new job Queue
//and assign all required parameters
func NewQueue(workers int, capacity int, jobCallBack JobCallBack) Queue {
	var wg sync.WaitGroup
	jobQueue := make(chan interface{}, capacity)
	quit := make(chan struct{})

	return Queue{
		Workers:     workers,
		JobQueue:    jobQueue,
		JobCallBack: jobCallBack,
		Wg:          &wg,
		QuitChan:    quit,
	}

}

// Stop close all the running goroutines
// and stops processing any more jobs
func (q *Queue) Stop() {
	q.QuitChan <- struct{}{}

}

//EnqueueJobNonBlocking  use this to queue the jobs you need to execute
//in an unblocking way...(i.e if the JobQueue is full it will not block)
//Returns false if the buffer is full
//else if it is accepted the job it returns true
//use case imagine you are receiving jobs and you want to prevent any more
//jobs from being submitted if the buffered channel
//is full..you can return an error to the user if this function returns false...
//although a better approach would be to store it in redis is it is rejected for
//later processing once the JobQueue has available space
//Note  if you are using a for loop to consume the jobs, its better  to use [Queue.EnqueueJobBlocking ]
//to prevent you from having a Busy wait(continuous pooling to check if a space is available to queue the job)
//for example consider this program
//func ShowBusyWait() {
//	index := 0
//	done := make(chan bool)
//	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	queue := taskq.NewQueue(1, 5, func(job interface{}) {
//		time.Sleep(10 * time.Second)
//
//		if index == len(items) {
//			done <- true
//		}
//		fmt.Println("i have finished processing item item:", job)
//	})
//	go queue.StartWorkers()
//
//out:
//	for {
//
//		for index < len(items) {
//			item := items[index]
//			queued := queue.EnqueueJobNonBlocking(item)
//			if queued {
//				log.Println("item added was", item)
//				index += 1
//				if index == len(items) {
//
//					//finished queueing all items break from outer loop
//					break out
//				}
//			} else {
//				log.Println("item was not queued unnecessary loop here")
//
//			}
//
//		}
//
//	}
//	//to wait for workers to finish
//	<-done
//
//	//time.Sleep(50 * time.Second)
//}
//since  the job is taking to long to be processed(delay of 10 second)
//our program will continue looping until they is an empty slot for  to queue a job
//this can be fixed by blocking the channel if its full which is what [Queue.EnqueueJobBlocking] does
func (q *Queue) EnqueueJobNonBlocking(job interface{}) bool {
	select {
	case q.JobQueue <- job:
		return true
	default:
		return false
	}
}

// EnqueueJobBlocking queues jobs and blocks if JobQueue is full
//once the JobQueue is no longer full the job will be accepted
//this is better for your cpu utilization unlike [Queue.EnqueueJobNonBlocking]
//when consuming a job via a for loop
func (q *Queue) EnqueueJobBlocking(job interface{}) {

	q.JobQueue <- job
}

// StartWorkers start  goroutines  and add them to wait group
//the goroutines added will be determined by the number of [Queue.Workers] you specified for this queue
//for example if you specified 10 Workers 10 goroutines will be used to process the job
//they can now start picking jobs and processing them
func (q *Queue) StartWorkers() {
	for i := 0; i < q.Workers; i++ {
		//add the goroutine  to a wait group to prevent the program from exiting
		//be a goroutine return
		q.Wg.Add(1)
		go q.worker()
	}
	q.Wg.Wait()
}

//worker is the function that will be run by each goroutine
//The JobCallBack  function specifies what should happen to each payload received by each worker
//Note each JobCallBack should implement yuri.JobCallBack i.e should have a process method
//The process method is what will be called on each job received
//For example imagine your workers are receiving numbers and you wanted to check if the number is even
//this is how you would go about it:
//Note you could user a high order function instead of creating a method on struct
//this is basically to show you how you can pass other dependency you may need
//such as redis or database
//type NumberIsEvenQueue struct {
//other dependencies go here
//such as database
//or redis

//}
//
//func (N NumberIsEvenQueue) JobCallBack(item interface{}) {
//	//check if item is an integer first
//	if n, ok := item.(int); ok {
//		if n%2 == 0 {
//			fmt.Printf("%v is an even number\n", item)
//		} else {
//			fmt.Printf("%v is an odd number\n", item)
//		}
//
//	} else {
//		fmt.Printf("%v is not an integer\n", item)
//	}
//
//}
//func main() {
//	nonBlocking := NumberIsEvenQueue{}
//	queue := yuri.NewQueue(1, 5, nonBlocking.JobCallBack)
//	go queue.StartWorkers()
//	for i := 1; i <= 10; i++ {
//		queue.EnqueueJobBlocking(i)
//	}
//
//
//
//	time.Sleep(10 * time.Second)
//}
func (q *Queue) worker() {
	defer q.Wg.Done()
	for {
		select {
		//terminate the goroutine
		case <-q.QuitChan:
			log.Println("closing the  workers")
			return

		case job := <-q.JobQueue:
			//a job has been received  call this function
			q.JobCallBack(job)
		}
	}
}
