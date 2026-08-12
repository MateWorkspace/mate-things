package subscriptions

import (
	"context"

	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurenodeshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/node/shared"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	registrationQos = 1
	statusQos       = 1
	logQos          = 0
	actionAckQos    = 1
	telemetryQos    = 1
	irRxQos         = 1
	irTxAckQos      = 1
)

type mqttImpl struct {
	client mqtt.Client
}

func NewMqttImpl(client mqtt.Client) domaincontractsnode.Subscriptions {
	return &mqttImpl{
		client: client,
	}
}

func (m *mqttImpl) Registration(ctx context.Context) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.GenericPubTopic("registration"), registrationQos)
}

func (m *mqttImpl) Status(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "status"), statusQos)
}

func (m *mqttImpl) Log(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "log"), logQos)
}

func (m *mqttImpl) ActionAck(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "action_ack"), actionAckQos)
}

func (m *mqttImpl) Telemetry(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "telemetry"), telemetryQos)
}

func (m *mqttImpl) IrCapture(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "ir/rx"), irRxQos)
}

func (m *mqttImpl) IrTransmitAck(ctx context.Context, nodeDeviceId string) (err error) {
	return m.subscribe(ctx, infrastructurenodeshared.NodePubTopic(nodeDeviceId, "ir/tx_ack"), irTxAckQos)
}

func (m *mqttImpl) subscribe(ctx context.Context, topic string, qos byte) error {
	token := m.client.Subscribe(topic, qos, nil)

	select {
	case <-ctx.Done():
		return domainmodels.NewError(
			"failed to publish due to context timeout",
			domainmodels.ErrTypeTimeout,
			ctx.Err(),
		)

	case <-token.Done():
		if err := token.Error(); err != nil {
			return domainmodels.NewError(
				"failed to publish due to token timeout",
				domainmodels.ErrTypeTimeout,
				err,
			)
		}
	}

	return nil
}
