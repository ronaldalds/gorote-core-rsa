package gorote

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type InitRabbitMQ struct {
	User     string
	Password string
	Host     string
	Port     int
	Vh       string
}

type ConnRabbitMQ struct {
	*amqp.Connection
	*amqp.Channel
}

type HandlesRabbitMQ func(context.Context, amqp.Delivery) error

func Redelivery(b amqp.Delivery) int {
	count, ok := b.Headers["x-delivery-count"]
	if !ok {
		return 0
	}
	reenvio, err := strconv.Atoi(fmt.Sprintf("%v", count))
	if err != nil {
		return 0
	}
	return reenvio
}

func (r *InitRabbitMQ) Connect(ctx context.Context) (*ConnRabbitMQ, error) {
	user := url.QueryEscape(r.User)
	pass := url.QueryEscape(r.Password)
	vhost := url.PathEscape(r.Vh)
	uri := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, pass, r.Host, r.Port, vhost)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal: %w", err)
	}

	return &ConnRabbitMQ{Connection: conn, Channel: ch}, nil
}

func (r *ConnRabbitMQ) ConsumerMessages(ctx context.Context, worker int, queue, nameConsumer string, handler HandlesRabbitMQ) error {
	if err := r.Channel.Qos(worker, 0, false); err != nil {
		return fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	msgs, err := r.Channel.Consume(queue, nameConsumer, false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("falha ao registrar consumidor: %w", err)
	}

	sem := make(chan struct{}, worker)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			if err := r.Channel.Close(); err != nil {
				log.Printf("Erro ao fechar canal: %v", err)
			}
			if err := r.Connection.Close(); err != nil {
				log.Printf("Erro ao fechar conexão: %v", err)
			}
			return fmt.Errorf("contexto encerrado. Finalizando RabbitMQ")
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("erro ao receber mensagens")
			}
			sem <- struct{}{}
			wg.Add(1)
			go func(msg amqp.Delivery) {
				defer func() {
					<-sem
					wg.Done()
				}()
				fmt.Println("[Worker] Processando nova mensagem...")
				if err := handler(ctx, msg); err != nil {
					d.Nack(false, true)
					return
				} else {
					d.Ack(false)
					return
				}
			}(d)
		}
	}
}
