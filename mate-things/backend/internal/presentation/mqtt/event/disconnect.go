package presentationmqttevent

import (
	"context"

	domainmodels "github.com/ABA-Developer/nusapala-things/backend/internal/domain/models"
	presentationmqtthandler "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/mqtt/handler"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func OnDisconnected(_ mqtt.Client, err error) {
	go func() {
		const tag = path + "/OnDisconnected"

		h := presentationmqtthandler.H
		if h == nil {
			return
		}

		h.Logger.Warn(context.Background(), tag, "mqtt disconnected", domainmodels.LoggerMeta{
			"err": err,
		})
	}()
}
