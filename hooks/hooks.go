package hooks

import (
	"bytes"
	"fmt"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type ExampleHookOptions struct {
	Server *mqtt.Server
}

type ExampleHook struct {
	mqtt.HookBase
	config *ExampleHookOptions
}

func (h *ExampleHook) ID() string {
	return "events-example"
}

func (h *ExampleHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnDisconnect,
		mqtt.OnSubscribed,
		mqtt.OnUnsubscribed,
		mqtt.OnPublish,
		mqtt.OnPublished,
	}, []byte{b})
}

func (h *ExampleHook) Init(config any) error {
	h.Log.Info("initialized")

	if _, ok := config.(*ExampleHookOptions); !ok && config != nil {
		return mqtt.ErrInvalidConfigType
	}

	h.config = config.(*ExampleHookOptions)
	if h.config.Server == nil {
		return mqtt.ErrInvalidConfigType
	}

	return nil
}

func (h *ExampleHook) subscribeCallbackfunc(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	h.Log.Info("hook subscribed message", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName)
}

func (h *ExampleHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	h.Log.Info("client connected", "client", cl.ID)

	h.config.Server.Subscribe("hook/direct/publish", 1, h.subscribeCallbackfunc)

	err := h.config.Server.Publish("hook/direct/publish", []byte("packet hook message"), false, 0)
	if err != nil {
		h.Log.Error("hook.publish", "error", err)
	}

	return nil
}

func (h *ExampleHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	if err != nil {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire, "error", err)
	} else {
		h.Log.Info("client disconnected", "client", cl.ID, "expire", expire)
	}
}

func (h *ExampleHook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	h.Log.Info(fmt.Sprintf("subscribed qos=%v", reasonCodes), "client", cl.ID, "filters", pk.Filters)
}

func (h *ExampleHook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("unsubscribed", "client", cl.ID, "filters", pk.Filters)
}

func (h *ExampleHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("received from client", "client", cl.ID, "payload", string(pk.Payload))

	pkx := pk
	if string(pk.Payload) == "hello" {
		pkx.Payload = []byte("hello world")
		h.Log.Info("received modified packet from client", "client", cl.ID, "payload", string(pkx.Payload))
	}

	return pkx, nil
}

func (h *ExampleHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	h.Log.Info("published to client", "client", cl.ID, "payload", string(pk.Payload))
}
