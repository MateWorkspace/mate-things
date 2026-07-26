package presentationmqttutils

import mqtt "github.com/eclipse/paho.mqtt.golang"

type msgcopy struct {
	duplicate bool
	qos       byte
	retained  bool
	topic     string
	messageId uint16
	payload   []byte
}

func Msgcopy(msg mqtt.Message) mqtt.Message {
	payload := append([]byte(nil), msg.Payload()...)
	return &msgcopy{
		duplicate: msg.Duplicate(),
		qos:       msg.Qos(),
		retained:  msg.Retained(),
		topic:     msg.Topic(),
		messageId: msg.MessageID(),
		payload:   payload,
	}
}

func (m *msgcopy) Duplicate() bool {
	return m.duplicate
}

func (m *msgcopy) Qos() byte {
	return m.qos
}

func (m *msgcopy) Retained() bool {
	return m.retained
}

func (m *msgcopy) Topic() string {
	return m.topic
}

func (m *msgcopy) MessageID() uint16 {
	return m.messageId
}

func (m *msgcopy) Payload() []byte {
	return m.payload
}

func (m *msgcopy) Ack() {}
