package greeter

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
)

func TestIPCExchangeUsesNativeLengthPrefixedJSON(t *testing.T) {
	clientConnection, serverConnection := net.Pipe()
	defer clientConnection.Close()
	defer serverConnection.Close()
	done := make(chan error, 1)
	go func() {
		var length [4]byte
		if _, err := io.ReadFull(serverConnection, length[:]); err != nil {
			done <- err
			return
		}
		payload := make([]byte, binary.NativeEndian.Uint32(length[:]))
		if _, err := io.ReadFull(serverConnection, payload); err != nil {
			done <- err
			return
		}
		var request Request
		if err := json.Unmarshal(payload, &request); err != nil {
			done <- err
			return
		}
		if request.Type != "create_session" || request.Username != "operator" {
			done <- io.ErrUnexpectedEOF
			return
		}
		response := []byte(`{"type":"auth_message","auth_message_type":"secret","auth_message":"Password:"}`)
		binary.NativeEndian.PutUint32(length[:], uint32(len(response)))
		if err := writeAll(serverConnection, length[:]); err != nil {
			done <- err
			return
		}
		done <- writeAll(serverConnection, response)
	}()
	response, err := NewClient(clientConnection).Exchange(Request{Type: "create_session", Username: "operator"})
	if err != nil || response.Type != "auth_message" || response.AuthMessageType != "secret" {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestIPCRejectsUnknownAndOversizedResponses(t *testing.T) {
	for _, response := range [][]byte{[]byte(`{"type":"invented"}`), nil} {
		clientConnection, serverConnection := net.Pipe()
		done := make(chan struct{})
		go func(payload []byte) {
			defer close(done)
			var requestLength [4]byte
			_, _ = io.ReadFull(serverConnection, requestLength[:])
			request := make([]byte, binary.NativeEndian.Uint32(requestLength[:]))
			_, _ = io.ReadFull(serverConnection, request)
			if payload == nil {
				binary.NativeEndian.PutUint32(requestLength[:], maxIPCMessage+1)
				_, _ = serverConnection.Write(requestLength[:])
			} else {
				binary.NativeEndian.PutUint32(requestLength[:], uint32(len(payload)))
				_, _ = serverConnection.Write(requestLength[:])
				_, _ = serverConnection.Write(payload)
			}
			_ = serverConnection.Close()
		}(response)
		_, err := NewClient(clientConnection).Exchange(Request{Type: "cancel_session"})
		if err == nil {
			t.Fatal("invalid response accepted")
		}
		_ = clientConnection.Close()
		<-done
	}
}
