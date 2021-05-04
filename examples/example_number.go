package examples

import "github.com/keithwachira/go-taskq"

// AddNumbers sample of how you can sum numbers
func AddNumbers() (int, int) {
	done := make(chan bool)
	initialValue := 100
	sum := 0
	q := taskq.NewQueue(1, 4, func(job interface{}) {
		if number, ok := job.(int); ok {
			sum += number
			initialValue -= number

			if number == 9 {
				//finished processing numbers
				done <- true
			}
		}
	})

	go q.StartWorkers()
	for i := 0; i < 10; i++ {
		q.EnqueueJobBlocking(i)
	}
	<-done
	return sum, initialValue

}
