package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"strings"
)

type WebsocketClient struct {
	Socket *websocket.Conn
	ahrs   *AHRSFilter
}

type WebsocketClientMessage struct {
	Type string
	Data interface{}
}

type MessageHandler func(client *WebsocketClient, msg *WebsocketClientMessage)

var handlerMap = map[string]MessageHandler{
	"ping":               Ping,
	"info":               Info,
	"reset_ahrs":         ResetAHRS,
	"subscribe_motion":   SubscribeMotion,
	"unsubscribe_motion": UnsubscribeMotion,
	"set_fps":            SetFPS,
}

func Ping(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.Send("pong", msg.Data)
}

func Info(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.Send("info", "HiFi Puppet Server")
}

func SetFPS(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.ahrs.UpdateSampleFrequency(float32(msg.Data.(float64)))
}

func ResetAHRS(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.ahrs.Reset()
}

func SubscribeMotion(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.ahrs.Subscribe(client)
}

func UnsubscribeMotion(client *WebsocketClient, msg *WebsocketClientMessage) {
	client.ahrs.Unsubscribe(client)
}

func UnknownMessageType(client *WebsocketClient, msg *WebsocketClientMessage) {
	fmt.Println("Unsupported message from client:", msg)
}

func handlerForClientMessage(msg *WebsocketClientMessage) MessageHandler {
	key := strings.ToLower(string(msg.Type))
	comm, ok := handlerMap[key]
	if !ok {
		return UnknownMessageType
	}
	return comm
}

func (self *WebsocketClient) Serve() {
	fmt.Println("Remote client connected.")
	for {
		var msg WebsocketClientMessage
		if err := websocket.JSON.Receive(self.Socket, &msg); err != nil {
			self.ahrs.Unsubscribe(self) // unsubscribe from motion
			fmt.Println("Remote client disconnected.")
			return
		}
		handlerForClientMessage(&msg)(self, &msg)
	}
}

func (self *WebsocketClient) Send(t string, data interface{}) {
	websocket.JSON.Send(self.Socket, &WebsocketClientMessage{t, data})
}

func (self *WebsocketClient) Motion(frame *AHRSQuaternionFrame) {
	self.Send("motion", frame)
}

func AHRSMotionServer(ahrs *AHRSFilter) websocket.Handler {
	return func(ws *websocket.Conn) {
		client := &WebsocketClient{
			Socket: ws,
			ahrs:   ahrs,
		}
		client.Serve()
	}
}
