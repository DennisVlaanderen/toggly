package fqdp

// FQDPMessageType represents the type of a FQDP message.
type FQDPMessageType uint8

const (
	// MessageTypeHandshake is the message type for handshake messages.
	MessageTypeHandshake     FQDPMessageType = 0x01
	// MessageTypeHandshakeAck is the message type for handshake acknowledgment messages.
	MessageTypeHandshakeAck  FQDPMessageType = 0x02
	// MessageTypeQuery is the message type for query messages.
	MessageTypeQuery         FQDPMessageType = 0x10
	// MessageTypeQueryResponse is the message type for query response messages.
	MessageTypeQueryResponse FQDPMessageType = 0x11
	// MessageTypeSubscribe is the message type for subscribe messages.
	MessageTypeSubscribe     FQDPMessageType = 0x20
	// MessageTypeUnsubscribe is the message type for unsubscribe messages.
	MessageTypeUnsubscribe   FQDPMessageType = 0x21
	// MessageTypeUpdate is the message type for update messages.
	MessageTypeUpdate        FQDPMessageType = 0x22
	// MessageTypeAck is the message type for acknowledgment messages.
	MessageTypeAck           FQDPMessageType = 0x30
	// MessageTypeNack is the message type for negative acknowledgment messages.
	MessageTypeNack          FQDPMessageType = 0x31
	// MessageTypeHeartbeat is the message type for heartbeat messages.
	MessageTypeHeartbeat     FQDPMessageType = 0x40
	// MessageTypeError is the message type for error messages.
	MessageTypeError         FQDPMessageType = 0x50
	// MessageTypeAdmin is the message type for admin messages.
	MessageTypeAdmin         FQDPMessageType = 0xF0
)
