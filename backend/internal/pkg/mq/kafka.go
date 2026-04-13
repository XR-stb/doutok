package mq

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/xiaoran/doutok/internal/pkg/logger"
)

// Topics
const (
	TopicVideoUpload   = "video.upload"
	TopicVideoProcess  = "video.process"
	TopicLikeEvent     = "social.like"
	TopicCommentEvent  = "social.comment"
	TopicFollowEvent   = "social.follow"
	TopicLiveGift      = "live.gift"
	TopicChatMessage   = "chat.message"
	TopicUserAction    = "analytics.user_action"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Send(ctx context.Context, topic, key string, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		logger.L().Errorw("kafka send failed", "topic", topic, "error", err)
	}
	return err
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) Read(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logger.L().Errorw("kafka read error", "error", err)
			continue
		}
		if err := handler(msg.Key, msg.Value); err != nil {
			logger.L().Errorw("message handler error", "error", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
