package mqtt

type Identifier interface {
	MQTTID() string
	Device() string
	Valid() bool
}
