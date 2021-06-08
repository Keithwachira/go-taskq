# A Extremely simple task (job) Queue for go

Go-taskq is a simple golang job queue that you can use with redis streams or any other event producer.

It is made simple to make sure anyone can easily customize it for their need.

## Quickstart
```go
// AddNumbers sample of how you can sum numbers
func AddNumbers() (int, int) {
	done := make(chan bool)
	complete:=0
	sum := 0
	var mutex = &sync.Mutex{}
	q := taskq.NewQueue(1, 4, func(job interface{}) {
		if number, ok := job.(int); ok {
			mutex.Lock()
			sum += number
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
	return sum

}
```

## How to use it with redis streams and  pass dependencies
```go
import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/keithwachira/go-taskq"
)

//this will show you how to pass dependencies such as redis

type RedisStreamsProcessing struct {
	Redis *redis.Client
    //you can also pass database here
    //or any other dependecies you have e.g loggger etc
}

// Process this method implements JobCallBack
///it will read and process redis streams
func (r *RedisStreamsProcessing) Process(job interface{}) {
	//do something with the stream received for example print it
	fmt.Println("Job received is", job)

}
func StartProcessingRedisStreams() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", 
		DB:       0,  
	})
	redisStreams := RedisStreamsProcessing{
		Redis: rdb,
	}

	q := taskq.NewQueue(5, 10, redisStreams.Process)
	//call startWorkers it in a different go routine otherwise it will block
	go q.StartWorkers()
	//add something to the redis sream
	data := map[string]interface{}{"tesd data": "data"}
	err := rdb.XAdd(context.Background(), &redis.XAddArgs{
		Stream:       "test_streams",
		MaxLen:       0,
		MaxLenApprox: 0,
		ID:           "",
		Values:       data,
	}).Err()
	if err != nil {
		log.Fatal(err)
	}

	id := "0"
	for {
		var ctx = context.Background()
		data, err := rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"test_streams", id},
			Count:   4,
			Block:   0,
		}).Result()
		if err != nil {
			log.Println(err)
			log.Fatal(err)
		}
		for _, result := range data {
			for _, message := range result.Messages {
				q.EnqueueJobBlocking(message)

				id = message.ID

			}

		}

	}

}


```
## Somethings to note
1. Always start  q.StartWorkers() in a new goroutine  i.e(go q.StartWorkers()) otherwise it will block.

2. If you are consumming you events in a for loop it is better to use **EnqueueJobBlocking** than **EnqueueJobNonBlocking** to prevent unnessesary cpu usage.
Consider the program below for instance:
```go
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
```
If you consumer is slow (like in this case where we are sleeping for 10 seconds before processing a job) your for loop will continue looping waiting
for a space in your job queue to open.Which will use unncecessary cpu.
If we used EnqueueJobBlocking it would have blocked until a space was open in our queue.


