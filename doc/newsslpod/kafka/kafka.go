package kafka

import (
	"fmt"
	"mysslee_qcloud/config"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var (
	testGlobal    int
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
)

// InitProducer 初始化producer
func InitProducer() {
	testGlobal = 111
	var err error
	KafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(config.Conf.Kafka.Servers, ","),
		// 用户不显示配置时，默认值为1。用户根据自己的业务情况进行设置
		"acks": 1,
		// 请求发生错误时重试次数，建议将该值设置为大于0，失败重试最大程度保证消息不丢失
		"retries": 0,
		// 发送请求失败时到下一次重试请求之间的时间
		"retry.backoff.ms": 100,
		// producer 网络请求的超时时间。
		"socket.timeout.ms": 6000,
		// 设置客户端内部重试间隔。
		"reconnect.backoff.max.ms": 3000,
		// 用户名密码
		"sasl.mechanism":    "PLAIN",
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.password":     config.Conf.Kafka.SASL.Password,
		"sasl.username":     config.Conf.Kafka.SASL.Username,
	})
	if err != nil {
		panic(err)
	}

	// 产生的消息 传递至报告处理程序
	go func() {
		for e := range KafkaProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logrus.Error(fmt.Sprintf("Delivery failed: %v\n", ev.TopicPartition))
				} else {
					logrus.Info(fmt.Sprintf("Delivered message to %v\n", ev.TopicPartition))
				}
			}
		}
	}()
}

// InitConsumer 初始化消费者
func InitConsumer() {
	var err error
	KafkaConsumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		// 设置接入点，请通过控制台获取对应Topic的接入点。
		"bootstrap.servers": strings.Join(config.Conf.Kafka.Servers, ","),
		// 设置的消息消费组
		"group.id":          config.Conf.Kafka.ConsumerGroupId,
		"auto.offset.reset": "latest",
		// 使用 Kafka 消费分组机制时，消费者超时时间。当 Broker 在该时间内没有收到消费者的心跳时，认为该消费者故障失败，Broker
		// 发起重新 Rebalance 过程。目前该值的配置必须在 Broker 配置group.min.session.timeout.ms=6000和group.max.session.timeout.ms=300000 之间
		"session.timeout.ms": 10000,
		// 用户名密码
		"sasl.mechanism":    "PLAIN",
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.password":     config.Conf.Kafka.SASL.Password,
		"sasl.username":     config.Conf.Kafka.SASL.Username,
	})
	if err != nil {
		panic(err)
	}
	err = KafkaConsumer.SubscribeTopics([]string{config.Conf.Kafka.Topic}, nil)
	if err != nil {
		panic(err)
	}
}

// ProduceMessage 产生消息
func ProduceMessage(message []byte) {
	KafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &config.Conf.Kafka.Topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)
}

// ConsumeMessage 消费信息
func ConsumeMessage() (*kafka.Message, error) {
	msg, err := KafkaConsumer.ReadMessage(-1)
	if err == nil {
		logrus.Info(fmt.Sprintf("Message on %s\n", msg.TopicPartition))
	} else {
		// 客户端将自动尝试恢复所有的 error
		logrus.Info(fmt.Sprintf("Consumer error: %v (%v)\n", err, msg))
	}
	return msg, err
}
