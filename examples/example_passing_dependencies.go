package examples

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
		Password: "", // no password set
		DB:       0,  // use default DB
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
