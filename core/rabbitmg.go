package core

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *InitRabbitMQ) StartRabbitConsumer(
	queue string,
	worker int,
	nameConsumer string,
	callback func(msg []byte) error,
) error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	user := url.QueryEscape(r.User)
	pass := url.QueryEscape(r.Password)
	vhost := url.PathEscape(r.Vh)
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, pass, r.Host, r.Port, vhost)
	conn, err := amqp.Dial(url)
	if err != nil {
		cancel()
		return fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		cancel()
		conn.Close()
		return fmt.Errorf("falha ao abrir canal: %w", err)
	}

	if err := ch.Qos(worker, 0, false); err != nil {
		cancel()
		conn.Close()
		return fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	msgs, err := ch.Consume(queue, nameConsumer, false, true, false, false, nil)
	if err != nil {
		cancel()
		ch.Close()
		conn.Close()
		return fmt.Errorf("falha ao registrar consumidor: %w", err)
	}

	go func() {
		defer func() {
			cancel()
			ch.Close()
			conn.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("Contexto encerrado. Finalizando consumo da fila.")
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("Canal de mensagens fechado pelo RabbitMQ.")
					return
				}
				if err := callback(d.Body); err != nil {
					log.Printf("Erro ao processar mensagem: %v", err)
					d.Nack(false, true)
					continue
				}
				d.Ack(false)
			}
		}
	}()

	return nil
}
