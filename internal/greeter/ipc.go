package greeter

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

const maxIPCMessage = 1024 * 1024

type Request struct {
	Type     string   `json:"type"`
	Username string   `json:"username,omitempty"`
	Response *string  `json:"response,omitempty"`
	Command  []string `json:"cmd,omitempty"`
	Env      []string `json:"env,omitempty"`
}

type Response struct {
	Type            string `json:"type"`
	ErrorType       string `json:"error_type,omitempty"`
	Description     string `json:"description,omitempty"`
	AuthMessageType string `json:"auth_message_type,omitempty"`
	AuthMessage     string `json:"auth_message,omitempty"`
}

type Client struct{ connection net.Conn }

func Dial(socket string) (*Client, error) {
	if socket == "" {
		return nil, fmt.Errorf("GREETD_SOCK is not set")
	}
	connection, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	return &Client{connection: connection}, nil
}

func NewClient(connection net.Conn) *Client { return &Client{connection: connection} }

func (c *Client) Close() error { return c.connection.Close() }

func (c *Client) Exchange(request Request) (Response, error) {
	payload, err := marshalRequest(request)
	if err != nil {
		return Response{}, err
	}
	if len(payload) > maxIPCMessage {
		return Response{}, fmt.Errorf("greetd IPC request exceeds size limit")
	}
	var length [4]byte
	binary.NativeEndian.PutUint32(length[:], uint32(len(payload)))
	if err := writeAll(c.connection, length[:]); err != nil {
		return Response{}, err
	}
	if err := writeAll(c.connection, payload); err != nil {
		return Response{}, err
	}
	if _, err := io.ReadFull(c.connection, length[:]); err != nil {
		return Response{}, err
	}
	responseLength := binary.NativeEndian.Uint32(length[:])
	if responseLength == 0 || responseLength > maxIPCMessage {
		return Response{}, fmt.Errorf("invalid greetd IPC response length %d", responseLength)
	}
	responsePayload := make([]byte, responseLength)
	if _, err := io.ReadFull(c.connection, responsePayload); err != nil {
		return Response{}, err
	}
	var response Response
	if err := json.Unmarshal(responsePayload, &response); err != nil {
		return Response{}, fmt.Errorf("decode greetd IPC response: %w", err)
	}
	if response.Type != "success" && response.Type != "error" && response.Type != "auth_message" {
		return Response{}, fmt.Errorf("unknown greetd IPC response %q", response.Type)
	}
	return response, nil
}

func marshalRequest(request Request) ([]byte, error) {
	switch request.Type {
	case "create_session":
		if request.Username == "" {
			return nil, fmt.Errorf("greetd username is empty")
		}
		return json.Marshal(struct {
			Type     string `json:"type"`
			Username string `json:"username"`
		}{request.Type, request.Username})
	case "post_auth_message_response":
		return json.Marshal(struct {
			Type     string  `json:"type"`
			Response *string `json:"response"`
		}{request.Type, request.Response})
	case "start_session":
		if len(request.Command) == 0 {
			return nil, fmt.Errorf("greetd session command is empty")
		}
		return json.Marshal(struct {
			Type    string   `json:"type"`
			Command []string `json:"cmd"`
			Env     []string `json:"env"`
		}{request.Type, request.Command, request.Env})
	case "cancel_session":
		return []byte(`{"type":"cancel_session"}`), nil
	default:
		return nil, fmt.Errorf("unknown greetd IPC request %q", request.Type)
	}
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) != 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrUnexpectedEOF
		}
		data = data[written:]
	}
	return nil
}
