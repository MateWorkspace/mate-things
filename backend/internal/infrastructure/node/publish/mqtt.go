package infrastructurenodepublish

import (
	"context"
	"encoding/json"

	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	infrastructurenodeshared "github.com/MateWorkspace/mate-things/backend/internal/infrastructure/node/shared"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	registrationAckQos = 1
	otaQos             = 1
	actionQos          = 1
	configQos          = 1
	irTransmitQos      = 1
)

type mqttImpl struct {
	client mqtt.Client
}

func NewMqttImpl(client mqtt.Client) domaincontractsnode.Publish {
	return &mqttImpl{
		client: client,
	}
}

func (m *mqttImpl) RegistrationAck(
	ctx context.Context,
	nodeDeviceId string,
	success bool,
) (err error) {
	payload, err := json.Marshal(infrastructurenodeshared.RegistrationAckPayload{Success: success})
	if err != nil {
		return domainmodels.NewError("failed to build payload", domainmodels.ErrTypeValidation, err)
	}

	return m.publish(
		ctx, infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "registration_ack"),
		registrationAckQos, false, payload,
	)
}

func (m *mqttImpl) Ota(
	ctx context.Context,
	nodeDeviceId string,
	firmwareUrl string,
	firmwareSize int32,
	firmwareChecksum string,
) (err error) {
	payload, err := json.Marshal(infrastructurenodeshared.OtaPayload{
		FirmwareUrl:      firmwareUrl,
		FirmwareSize:     firmwareSize,
		FirmwareChecksum: firmwareChecksum,
	})
	if err != nil {
		return domainmodels.NewError("failed to build payload", domainmodels.ErrTypeValidation, err)
	}

	return m.publish(
		ctx, infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "ota"),
		otaQos, false, payload,
	)
}

func (m *mqttImpl) Action(
	ctx context.Context,
	nodeDeviceId string,
	executionId uuid.UUID,
	actionName string,
	payload json.RawMessage,
) (err error) {
	actionPayload, err := json.Marshal(infrastructurenodeshared.ActionPayload{
		ExecutionId: executionId,
		Action:      actionName,
		Payload:     payload,
	})
	if err != nil {
		return domainmodels.NewError("failed to build payload", domainmodels.ErrTypeValidation, err)
	}

	return m.publish(
		ctx, infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "action"),
		actionQos, false, actionPayload,
	)
}

func (m *mqttImpl) Config(
	ctx context.Context,
	nodeDeviceId string,
	key string,
	value string,
) (err error) {
	payload, err := json.Marshal(infrastructurenodeshared.ConfigPayload{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return domainmodels.NewError("failed to build payload", domainmodels.ErrTypeValidation, err)
	}

	return m.publish(
		ctx, infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "config"),
		configQos, false, payload,
	)
}

func (m *mqttImpl) IrTransmit(
	ctx context.Context,
	nodeDeviceId string,
	executionId uuid.UUID,
	rawData []int32,
) (err error) {
	payload, err := json.Marshal(infrastructurenodeshared.IrTransmitPayload{
		ExecutionId: executionId,
		RawData:     rawData,
	})
	if err != nil {
		return domainmodels.NewError("failed to build payload", domainmodels.ErrTypeValidation, err)
	}

	return m.publish(
		ctx, infrastructurenodeshared.NodeSubTopic(nodeDeviceId, "ir_transmit"),
		irTransmitQos, false, payload,
	)
}

func (m *mqttImpl) publish(ctx context.Context, topic string, qos byte, retained bool, payload []byte) error {
	token := m.client.Publish(topic, qos, retained, payload)

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
