package amqputil

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

type AMQPConn struct {
	Conn     *amqp.Connection
	dialFn   func() (*AMQPConn, error)
	attempts uint8
}

func New() *AMQPConn {
	return &AMQPConn{}
}

func (conn *AMQPConn) Dial(uri string) (*AMQPConn, error) {
	conn.dialFn = func() (*AMQPConn, error) {
		var err error

		conn.Conn, err = amqp.Dial(uri)

		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	return conn.dialFn()
}

func (conn *AMQPConn) AutoRedial() *AMQPConn {
	errChan := conn.Conn.NotifyClose(make(chan *amqp.Error))

	go func() {
		var err error

		select {
		case amqpErr := <-errChan:
			err = amqpErr
		attempt:
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}

			if conn.attempts > 60 {
				conn.attempts = 0
			}

			time.Sleep(time.Duration(int64(conn.attempts) * int64(time.Second)))

			conn, err = conn.dialFn()

			if err != nil {
				conn.attempts += 1
				goto attempt
			}
		}
	}()

	return conn
}
