package rabbitmq

import (
	"context"
	"fmt"
	"github.com/khanhnguyen02311/EchoChat-WS/components/manager"
	conf "github.com/khanhnguyen02311/EchoChat-WS/configurations"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type RMQService struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRMQService() (*RMQService, error) {
	conn, err := amqp.Dial("amqp://" + conf.PROTO_RMQ_USR + ":" + conf.PROTO_RMQ_PWD + "@" + conf.PROTO_RMQ_HOST + ":" + conf.PROTO_RMQ_PORT)
	if err != nil {
		fmt.Println("Error connecting to RabbitMQ: ", err.Error())
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Error opening channel: ", err.Error())
		return nil, err
	}
	// make sure queues exist
	_, err = ch.QueueDeclare(
		conf.PROTO_RMQ_QUEUE_NOTI, // name
		true,                      // durables
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		fmt.Println("Error declaring queue: ", err.Error())
		return nil, err
	}
	_, err = ch.QueueDeclare(
		conf.PROTO_RMQ_QUEUE_MSG, // name
		true,                     // durables
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		fmt.Println("Error declaring queue: ", err.Error())
		return nil, err
	}
	return &RMQService{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (r *RMQService) Close() {
	r.Channel.Close()
	r.Conn.Close()
}

func (r *RMQService) StartConsuming(ctx context.Context, manager *manager.ConnectionManager, wg *sync.WaitGroup) {
	messagesFromQueue, err := r.Channel.Consume(
		conf.PROTO_RMQ_QUEUE_MSG, // queue
		"",                       // consumer
		true,                     // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-waitc
		nil,                      // args
	)
	if err != nil {
		fmt.Println("Error consuming from queue: ", err.Error())
		return
	}
	notisFromQueue, err := r.Channel.Consume(
		conf.PROTO_RMQ_QUEUE_NOTI, // queue
		"",                        // consumer
		true,                      // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-waitc
		nil,                       // args
	)
	if err != nil {
		fmt.Println("Error consuming from queue: ", err.Error())
		return
	}
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done(): // exit when context is canceled
				return
			case d := <-messagesFromQueue:
				manager.ProcessRMQMessage(d.Body)
			case d := <-notisFromQueue:
				manager.ProcessRMQNoti(d.Body)
			}
		}
	}()
}

//func main() {
//	queuename := username + "$" + device
//	fmt.Printf("User name: %s\n", username)
//	fmt.Printf("Queue name: %s\n", queuename)
//
//	// conn, err := amqp.DialTLS("amqps://localhost:5671/", cfg)
//	conn, err := amqp.Dial("amqp://" + username + ":" + username + "@rabbit-testbench.duckdns.org:5672/testvhost")
//	failOnError(err, "Failed to connect to RabbitMQ")
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	failOnError(err, "Failed to open a channel")
//	defer ch.Close()
//
//	msgs, err := ch.Consume(
//		queuename, // queue
//		"",        // consumer
//		true,      // auto-ack
//		false,     // exclusive
//		false,     // no-local
//		false,     // no-waitc
//		nil,       // args
//	)
//	failOnError(err, "Failed to register a consumer")
//
//	var forever chan struct{}
//
//	go func() {
//		for d := range msgs {
//			log.Printf("Received a message: %s", d.Body)
//			// d.Ack(false)  // when consumer auto-ack = false
//		}
//	}()
//
//	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
//	<-forever
//}
//
//func setup() {
//	with_account_creation := true
//	exectype, username, device := bodyFrom(os.Args)
//
//	// conn, err := amqp.DialTLS("amqps://localhost:5671/", cfg)
//	conn, err := amqp.Dial("amqp://systemuser:systemuser@rabbit-testbench.duckdns.org:5672/testvhost")
//	failOnError(err, "Failed to connect to RabbitMQ")
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	failOnError(err, "Failed to open a channel")
//	defer ch.Close()
//
//	if exectype == "init" {
//		err = ch.ExchangeDeclare(
//			"system_global_exchange_01", // name
//			"fanout",                    // type
//			true,                        // durable
//			false,                       // auto-deleted
//			false,                       // internal
//			false,                       // no-wait
//			nil,                         // arguments
//		)
//		failOnError(err, "Failed to declare global exchange")
//
//		err = ch.ExchangeDeclare(
//			"system_direct_exchange_01", // name
//			"direct",                    // type
//			true,                        // durable
//			false,                       // auto-deleted
//			false,                       // internal
//			false,                       // no-wait
//			nil,                         // arguments
//		)
//		failOnError(err, "Failed to declare direct exchange")
//
//	} else if exectype == "user" {
//		queuename := username + "$" + device
//		fmt.Printf("User name: %s\n", username)
//		fmt.Printf("Queue name: %s\n", queuename)
//
//		if with_account_creation {
//			client := &http.Client{}
//			// create new user
//			url := "http://rabbit-testbench.duckdns.org:15672/api/users/" + username
//			payload := `{"password":"` + username + `","tags":""}`
//			req := createRequest("PUT", url, payload)
//			resp, err := client.Do(req)
//			failOnError(err, "Failed to create user")
//			fmt.Println("Create user Status:", resp.Status)
//			// defer resp.Body.Close()
//			// fmt.Println("response Status:", resp.Status)
//			// body, _ := ioutil.ReadAll(resp.Body)
//			// fmt.Println("response Body:", string(body))
//
//			// grant permissions to testvhost
//			url = "http://rabbit-testbench.duckdns.org:15672/api/permissions/testvhost/" + username
//			payload = `{"configure":"","write":"^\\Qsystem_direct\\E.*","read":"^\\Q` + username + `\\E.*"}`
//			req = createRequest("PUT", url, payload)
//			resp, err = client.Do(req)
//			fmt.Println("Grant vhost permission Status:", resp.Status)
//			failOnError(err, "Failed to grant permissions to vhost")
//		}
//
//		ch, err := conn.Channel()
//		failOnError(err, "Failed to open a channel")
//		defer ch.Close()
//
//		// create queue if not exist
//		q, err := ch.QueueDeclare(
//			queuename, // name
//			false,     // durables
//			false,     // delete when unused
//			false,     // exclusive
//			false,     // no-wait
//			nil,       // arguments
//		)
//		_ = q
//		failOnError(err, "Failed to declare queue")
//
//		// bind to global exchange
//		err = ch.QueueBind(
//			queuename,                   // queue name
//			"",                          // routing key
//			"system_global_exchange_01", // exchange
//			false,
//			nil,
//		)
//		failOnError(err, "Failed to bind queue")
//
//		// bind to direct exchange
//		err = ch.QueueBind(
//			queuename,                   // queue name
//			username,                    // routing key
//			"system_direct_exchange_01", // exchange
//			false,
//			nil,
//		)
//		failOnError(err, "Failed to bind queue")
//	}
//	fmt.Printf("Done\n")
//}
