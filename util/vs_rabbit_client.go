package util

import (
	"context"
	"github.com/Yordroid/rabbitmq"
	log "github.com/sirupsen/logrus"
)

type VsRabbitMQClient struct {
	rabbitCtx   *rabbitmq.Rabbitmq
	exchangeMap map[string]rabbitmq.ExchangeOptions
}

// VsOnMessageCallback exchangeName 对应kafka topic
type VsOnMessageCallback func(exchangeName string, msg []byte)
type rabbitMQOption struct {
	broken        string              //远程服务器信息
	exchangeNames []string            //交换机名称
	isDurable     bool                //是否持久化消息,只有生产者需要
	cbMessage     VsOnMessageCallback //消息回调
	queueName     string              //队列名称
	isProducer    bool                //是否为生产者,否则为消费者
	exchangeType  string
}
type VsRabbitMQOption func(*rabbitMQOption)

func WithExchangeType(exchangeType string) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.exchangeType = exchangeType
	}
}

func WithBrokenInfo(broken string) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.broken = broken
	}
}
func WithExchangeNames(exchangeNames []string) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.exchangeNames = exchangeNames
	}
}
func WithQueueName(queueName string) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		if queueName != "" {
			option.queueName = queueName
		}
	}
}
func WithDurable(isDurable bool) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.isDurable = isDurable
	}
}
func WithOnMessage(cbMessage VsOnMessageCallback) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.cbMessage = cbMessage
	}
}
func WithIsProducer(isProducer bool) VsRabbitMQOption {
	return func(option *rabbitMQOption) {
		option.isProducer = isProducer
	}
}

// InitClient 生产者决定exchange 是否持久化,消费者决定queue是否持久化
func (_self *VsRabbitMQClient) InitClient(options ...VsRabbitMQOption) {
	_self.exchangeMap = make(map[string]rabbitmq.ExchangeOptions)
	op := rabbitMQOption{
		broken:        "guest:guest@localhost:5672",
		exchangeNames: []string{"defaultExchange"},
		isDurable:     false,
		cbMessage:     nil,
		queueName:     "default",
		isProducer:    true,
		exchangeType:  "fanout",
	}
	// 调用动态传入的参数进行设置值
	for _, option := range options {
		option(&op)
	}

	r := rabbitmq.NewRabbitmq(
		op.broken,
		nil,
	)
	// 2. Create a connection to the rabbitmq service
	if err := r.Connect(); err != nil {
		log.Error("init rabbit mq fail:", err.Error())
		return
	}

	for idx := 0; idx < len(op.exchangeNames); idx++ {
		var ops []rabbitmq.Option

		// 3. Exchange configuration
		curExchangeName := op.exchangeNames[idx]

		// Set the exchange that the queue needs to bind to
		exchange := rabbitmq.ExchangeOptions{
			Name:    curExchangeName,                        // Exchange name
			Type:    rabbitmq.ExchangeType(op.exchangeType), // Exchange type
			Durable: op.isDurable,                           // Whether it is durable

		}
		if op.isProducer {
			err := r.Exchange(exchange)
			if err != nil {
				log.Error("InitClient Producer fail", err.Error())
				return
			}
			_self.exchangeMap[curExchangeName] = exchange
		} else {
			// Set the name of the consumption queue
			ops = append(ops, rabbitmq.Queue(op.queueName), rabbitmq.Exchange(exchange))
			if op.isDurable {
				ops = append(ops, rabbitmq.DurableQueue())
			}
			// 4. Create a subscription
			// Subscribing and consuming internally starts a goroutine that consumes data so it won't block the main goroutine.
			err := r.Subscribe(func(exchangeName string, msg []byte) error {
				if op.cbMessage == nil {
					log.Error("InitClient fail,not set WithOnMessage option")
					return nil
				}
				op.cbMessage(exchangeName, msg)
				return nil
			}, ops...)
			if err != nil {
				log.Error("InitClient fail,sub:", err.Error())
				return
			}
		}

	}
	_self.rabbitCtx = r
	log.Info("rabbit success")
}

// PublicMessage 发布消息
func (_self *VsRabbitMQClient) PublicMessage(exchangeName string, msg []byte) {
	if _self.rabbitCtx != nil {
		exchange, isExist := _self.exchangeMap[exchangeName]
		if !isExist {
			log.Error("publicMessage fail,not set exchange:", exchangeName)
			return
		}
		err := _self.rabbitCtx.Publish(context.Background(), msg, rabbitmq.Exchange(exchange))
		if err != nil {
			log.Error("publicMessage fail", err.Error())
			return
		}
	}
}
