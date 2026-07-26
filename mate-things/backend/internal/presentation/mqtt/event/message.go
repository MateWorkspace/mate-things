package presentationmqttevent

import (
	"context"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqtthandler "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/handler"
	presentationmqttutils "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type topicParts struct {
	Direction string
	DeviceId  string
	Suffix    string
	IsGlobal  bool
}

func OnMessage(_ mqtt.Client, msg mqtt.Message) {
	go func(msg mqtt.Message) {
		const tag = path + "/OnMessage"

		h := presentationmqtthandler.H
		if h == nil {
			return
		}

		ctx := context.Background()
		topic, ok := extractTopicParts(msg.Topic())
		if !ok || topic.Direction != "pub" {
			h.Logger.Warn(ctx, tag, "invalid mqtt topic", domainmodels.LoggerMeta{
				"topic": msg.Topic(),
			})
			return
		}

		if topic.IsGlobal {
			switch topic.Suffix {
			case "registration":
				h.Registration(ctx, msg)
			default:
				h.Logger.Warn(ctx, tag, "unknown mqtt topic", domainmodels.LoggerMeta{
					"topic": msg.Topic(),
				})
			}
			return
		}

		switch topic.Suffix {
		case "action_ack":
			h.ActionAck(ctx, msg, topic.DeviceId)
		case "status":
			h.Status(ctx, msg, topic.DeviceId)
		case "log":
			h.Log(ctx, msg, topic.DeviceId)
		default:
			h.Logger.Warn(ctx, tag, "unknown mqtt topic", domainmodels.LoggerMeta{
				"topic": msg.Topic(),
			})
		}
	}(presentationmqttutils.Msgcopy(msg))
}

func extractTopicParts(topic string) (topicParts, bool) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	switch len(parts) {
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return topicParts{}, false
		}
		return topicParts{
			Direction: parts[0],
			Suffix:    parts[1],
			IsGlobal:  true,
		}, true

	case 3:
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return topicParts{}, false
		}
		return topicParts{
			Direction: parts[0],
			DeviceId:  parts[1],
			Suffix:    parts[2],
		}, true

	default:
		return topicParts{}, false
	}
}
