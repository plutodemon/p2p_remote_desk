package main

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
)

func Test_SignalingServerBasic(t *testing.T) {
	// 创建两个WebSocket客户端
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url1 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=test_room"}
	url2 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=test_room"}

	c1, _, err := websocket.Dial(ctx, url1.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	c2, _, err := websocket.Dial(ctx, url2.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	// 测试消息转发
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "Hello from client 1",
	}

	// 客户端1发送消息
	msgBytes, err := marshalJSON(testMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	err = c1.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 设置读取超时
	readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
	defer readCancel()

	// 客户端2接收消息
	type readResult struct {
		messageType int
		message     []byte
		err         error
	}

	ch := make(chan readResult)
	go func() {
		msgType, msg, err := c2.ReadMessage()
		ch <- readResult{msgType, msg, err}
	}()

	select {
	case <-readCtx.Done():
		t.Fatal("Timeout waiting for message")
	case result := <-ch:
		if result.err != nil {
			t.Fatalf("Failed to read message: %v", result.err)
		}

		var receivedMsg map[string]interface{}
		err = unmarshalJSON(result.message, &receivedMsg)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// 验证消息内容
		assert.Equal(t, testMessage["type"], receivedMsg["type"])
		assert.Equal(t, testMessage["data"], receivedMsg["data"])
		assert.Equal(t, "test_room", receivedMsg["roomId"])
	}
}

func Test_SignalingServerMultipleRooms(t *testing.T) {
	// 创建两个不同房间的WebSocket客户端
	ctx := context.Background()
	url1 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=room1"}
	url2 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=room2"}

	c1, _, err := websocket.Dial(ctx, url1.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}
	defer c1.Close()

	c2, _, err := websocket.Dial(ctx, url2.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}
	defer c2.Close()

	// 测试消息不会跨房间转发
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "This message should not be received",
	}

	// 客户端1发送消息
	msgBytes, err := marshalJSON(testMessage)
	err = c1.Write(ctx, websocket.MessageText, msgBytes)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 设置带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// 客户端2不应该收到消息
	_, _, err = c2.Read(timeoutCtx)
	assert.Error(t, err) // 应该超时
}

func Test_SignalingServerDisconnection(t *testing.T) {
	// 创建两个WebSocket客户端
	ctx := context.Background()
	url1 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=test_room"}
	url2 := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=test_room"}

	c1, _, err := websocket.Dial(ctx, url1.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}

	c2, _, err := websocket.Dial(ctx, url2.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}
	defer c2.Close()

	// 关闭客户端1
	c1.Close(websocket.StatusNormalClosure, "")

	// 等待一段时间确保服务器处理了断开连接
	time.Sleep(1 * time.Second)

	// 客户端2发送消息
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "Message after disconnection",
	}

	msgBytes, err := marshalJSON(testMessage)
	err = c2.Write(ctx, websocket.MessageText, msgBytes)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
}
