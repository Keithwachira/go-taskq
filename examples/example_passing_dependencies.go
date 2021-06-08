package examples

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/keithwachira/go-taskq"
)

var streamName = "send_order_emails"

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // use default DB
	})

	handler := http.NewServeMux()
	///we create a new router to expose our api
	//to our users
	s := Server{Redis: rdb}
	handler.HandleFunc("/api/hello", s.NewOrderReceivedFromClient)
	//Every time a  request is sent to the endpoint ("/api/hello")
	//the function SayHello will be invoked
	err := http.ListenAndServe("0.0.0.0:8080", handler)
	if err != nil {
		log.Fatal(err)
	}
	//we tell our api to listen to all request to port 8080.
}

type Server struct {
	Redis *redis.Client
}

// NewOrderReceivedFromClient this is mocking an endpoint that users use to place an order
//once we receive an event here we should register it to redis
//then the workers will pick it up and process it
func (S *Server) NewOrderReceivedFromClient(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"email": "redis@gmail.com", "message": "We have received you order and we are working on it."}
	//we have received  an order here send it to
	//redis has a function called xadd that we will use to add this to our stream
	err := S.Redis.XAdd(context.Background(), &redis.XAddArgs{
		///this is the name we want to give to our stream
		///in our case we called it send_order_emails
		//note you can have as many stream as possible
		//such as one for email...another for notifications
		Stream:       streamName,
		MaxLen:       0,
		MaxLenApprox: 0,
		ID:           "",
		//values is the data you want to send to the stream
		//in our case we send a map with email and message
		Values: data,
	}).Err()
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `We received you order`)
}

// RedisStreamsProcessing this will show you how to pass dependencies such as redis
//you can also pass a database here too
//if you need it to process you work
type RedisStreamsProcessing struct {
	Redis *redis.Client
	//other dependencies each logger database goes here
}

// Process this method implements JobCallBack
///it will read and process each email and send it to our users
//the logic to send the emails goes here
func (r *RedisStreamsProcessing) Process(job interface{}) {
	if data, ok := job.(redis.XMessage); ok {
		email := data.Values["email"].(string)
		message := data.Values["message"].(string)
		fmt.Printf("I am sending an email to the email  %vwith message:%v   \n ", email, message)
	} else {
		log.Println("wrong type of data sent")
	}

}

func StartProcessingEmails(*redis.Client) {

	//create and pass redis to our consumers
	redisStreams := RedisStreamsProcessing{
		Redis: rdb,
	}
	//in this case we have started 5 goroutines so at any moment we will
	//be sending a maximum of 5 emails.
	//you can adjust this parameters to increase or reduce
	q := taskq.NewQueue(5, 10, redisStreams.Process)
	//call startWorkers it in a different goroutine otherwise it will block
	go q.StartWorkers()
	//with our workers running now we can start listening to new events from redis stream
	//and queue them

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
