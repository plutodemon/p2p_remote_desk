package main

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
)

func Test_SignalingServerConcurrentConnections(t *testing.T) {
	// 测试多个客户端并发连接
	const numClients = 10
	clients := make([]*websocket.Conn, numClients)
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ul := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=concurrent_test"}
			c, _, err := websocket.Dial(ctx, ul.String(), nil)
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", index, err)
				return
			}
			clients[index] = c
		}(i)
	}

	wg.Wait()

	// 验证连接成功数量
	successfulConnections := 0
	for _, client := range clients {
		if client != nil {
			successfulConnections++
			defer client.Close(websocket.StatusNormalClosure, "")
		}
	}

	assert.True(t, successfulConnections > 0, "至少应该有一个成功的连接")
}

func Test_SignalingServerMessageBroadcast(t *testing.T) {
	// 测试广播消息到多个客户端
	const numClients = 5
	clients := make([]*websocket.Conn, numClients)
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建多个客户端
	for i := 0; i < numClients; i++ {
		url := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=broadcast_test"}
		c, _, err := websocket.Dial(ctx, url.String(), nil)
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		clients[i] = c
		defer c.Close(websocket.StatusNormalClosure, "")
	}

	// 从第一个客户端发送消息
	testMessage := map[string]interface{}{
		"type": "broadcast",
		"data": "Broadcast message",
	}

	msgBytes, err := marshalJSON(testMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	err = clients[0].Write(ctx, websocket.MessageText, msgBytes)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 其他客户端接收消息
	for i := 1; i < numClients; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 设置消息接收超时
			readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
			defer readCancel()

			ch := make(chan struct {
				msg []byte
				err error
			})

			go func() {
				_, msg, err := clients[index].Read(readCtx)
				ch <- struct {
					msg []byte
					err error
				}{msg, err}
			}()

			select {
			case <-readCtx.Done():
				t.Errorf("Client %d timeout waiting for message", index)
				return
			case result := <-ch:
				if result.err != nil {
					t.Errorf("Client %d failed to read message: %v", index, result.err)
					return
				}

				var receivedMsg map[string]interface{}
				err = unmarshalJSON(result.msg, &receivedMsg)
				if err != nil {
					t.Errorf("Client %d failed to unmarshal message: %v", index, err)
					return
				}

				assert.Equal(t, testMessage["type"], receivedMsg["type"])
				assert.Equal(t, testMessage["data"], receivedMsg["data"])
			}
		}(i)
	}

	wg.Wait()
}

func Test_SignalingServerErrorHandling(t *testing.T) {
	// 测试错误处理场景
	ctx := context.Background()

	// 测试无效的房间ID
	invalidURL := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room="}
	c, _, err := websocket.Dial(ctx, invalidURL.String(), nil)
	if err == nil {
		c.Close(websocket.StatusNormalClosure, "")
		t.Error("Expected error for invalid room ID")
	}

	// 测试无效的消息格式
	validURL := url.URL{Scheme: "ws", Host: "localhost:28080", Path: "/ws", RawQuery: "room=error_test"}
	c, _, err = websocket.Dial(ctx, validURL.String(), nil)
	assert.NoError(t, err)
	defer c.Close(websocket.StatusNormalClosure, "")

	// 发送无效的JSON消息
	err = c.Write(ctx, websocket.MessageText, []byte("invalid json"))
	if err != nil {
		t.Fatalf("Failed to write invalid message: %v", err)
	}

	// 等待一段时间确保服务器没有崩溃
	time.Sleep(1 * time.Second)

	// 验证服务器仍然可以接收有效消息
	validMessage := map[string]interface{}{
		"type": "test",
		"data": "Valid message after invalid one",
	}
	msgBytes, err := marshalJSON(validMessage)
	err = c.Write(ctx, websocket.MessageText, msgBytes)
	if err != nil {
		t.Fatalf("Failed to write valid message: %v", err)
	}
}

// 辅助函数：JSON编码
func marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// 辅助函数：JSON解码
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
