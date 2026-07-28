package presentationmqttevent

import (
	"context"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationmqtthandler "github.com/MateWorkspace/mate-things/backend/internal/presentation/mqtt/handler"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func OnConnect(_ mqtt.Client) {
	go func() {
		const tag = path + "/OnConnect"

		var h *presentationmqtthandler.Handler
		for {
			h = presentationmqtthandler.H
			if h != nil {
				break
			}
			time.Sleep(time.Second)
		}

		ctx := context.Background()
		if err := h.MessagingCallback.Resubscribe(ctx); err != nil {
			h.Logger.Error(ctx, tag, "failed to resubscribe mqtt topics", domainmodels.LoggerMeta{
				"err": err,
			})
			return
		}
	}()
}
