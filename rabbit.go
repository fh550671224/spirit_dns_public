package dns

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const (
	SpiritDNSLog        = "spirit_dns_log"
	SpiritDNSBackendLog = "spirit_dns_backend_log"
)

type RabbitClient struct {
	*amqp.Connection
	*amqp.Channel
}

func (client *RabbitClient) IsOk() bool {
	return client.Channel != nil && client.Connection != nil
}

func (client *RabbitClient) Init(username string, password string, addr string) error {
	conn, err := amqp.Dial("amqp://" + username + ":" + password + "@" + addr + "/")
	if err != nil {
		return fmt.Errorf("connect to rabbit err:%v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("creating channel err:%v", err)
	}

	client.Channel = ch
	client.Connection = conn

	return nil
}

func (client *RabbitClient) CloseRabbit() {
	if !client.IsOk() {
		return
	}

	err := client.Channel.Close()
	if err != nil {
		log.Printf("RabbitMQ Channel Close err:%v", err)
	}
	err = client.Connection.Close()
	if err != nil {
		log.Printf("RabbitMQ Connection Close err:%v", err)
	}
}

func (client *RabbitClient) Write(queueName string, msg []byte) error {
	if !client.IsOk() {
		return fmt.Errorf("client is nil")
	}

	q, err := client.QueueDeclare(queueName, false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("QueueDeclare err:%v", err)
	}

	err = client.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msg,
	})
	if err != nil {
		return fmt.Errorf("Publish err:%v", err)
	}

	return nil
}
