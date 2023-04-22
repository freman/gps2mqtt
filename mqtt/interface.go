package mqtt

type Identifier interface {
	MQTTID() string
	Device() string
}

type Message struct {
	Type MessageType
	Data Identifier
}

type MessageType byte

const (
	TypeHello MessageType = iota
	TypeAlarm
	TypeUpdate
)
