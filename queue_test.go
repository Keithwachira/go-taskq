package taskq

import (
	"log"
	"testing"
)

func TestQueue_EnqueueJobNonBlocking(t *testing.T) {
	rejectedJobs := []int{}
	q := NewQueue(1, 4, func(job interface{}) {
		log.Print("received item here")

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
