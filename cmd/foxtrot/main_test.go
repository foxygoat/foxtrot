package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestSocketHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(socketHandler))
	defer server.Close()
	dialer := websocket.Dialer{}

	conn, resp, err := dialer.Dial("ws://"+server.Listener.Addr().String()+"/ws", nil)
	defer resp.Body.Close() //nolint: errcheck
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	message := []byte("echo test")
	err = conn.WriteMessage(websocket.TextMessage, message)
	require.NoError(t, err)

	_, got, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, message, got)
}
