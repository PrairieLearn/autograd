package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/PrairieLearn/autograd/grader"
)

const (
	consumerTag = "autograd-consumer"
)

type Client struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	gradingQueue amqp.Queue
	resultQueue  amqp.Queue
	grader       *grader.Grader
	done         chan error
}

func NewClient(amqpURI, gradingQueueName, resultQueueName string, grader *grader.Grader) (*Client, error) {
	c := &Client{
		conn:    nil,
		channel: nil,
		grader:  grader,
		done:    make(chan error),
	}

	var err error

	log.Debugf("Dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	log.Debugf("Got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}
	if err := c.channel.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("Channel Qos: %s", err)
	}

	log.Debugf("Got Channel, declaring Queues %q, %q", gradingQueueName, resultQueueName)
	c.gradingQueue, err = c.channel.QueueDeclare(gradingQueueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	c.resultQueue, err = c.channel.QueueDeclare(resultQueueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Debugf("Declared Queue (%q %d messages, %d consumers), starting Consume (consumer tag %q)",
		c.gradingQueue.Name, c.gradingQueue.Messages, c.gradingQueue.Consumers, consumerTag)
	deliveries, err := c.channel.Consume(c.gradingQueue.Name, consumerTag, false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go c.handle(deliveries, c.done)

	return c, nil
}

func (c *Client) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(consumerTag, true); err != nil {
		return fmt.Errorf("Client cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Debugf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func (c *Client) NotifyClose() chan *amqp.Error {
	return c.conn.NotifyClose(make(chan *amqp.Error))
}

func (c *Client) handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		log.WithFields(log.Fields{
			"queue":        c.gradingQueue.Name,
			"size":         len(d.Body),
			"delivery_tag": d.DeliveryTag,
		}).Info("Received grading job")
		log.Debug(string(d.Body))

		result, err := c.grader.Grade(d.Body)
		if err != nil {
			log.Warnf("Error initializing grader: %v", err)
			continue
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Warnf("Error marshalling grading result: %v", err)
			continue
		}

		msg := amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         resultJSON,
		}
		err = c.channel.Publish("", c.resultQueue.Name, false, false, msg)
		if err != nil {
			log.Warnf("Error publishing message to queue: %v", err)
		}

		d.Ack(false)
	}
	log.Debugf("handle: deliveries channel closed")
	done <- nil
}
