package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type LogFunc func(format string, args ...interface{})

func NewLogFunc() LogFunc {
	return func(format string, args ...interface{}) {
		log.Printf(format, args...)
	}
}

type jsonRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const ContentLength = "Content-Length: "

func readMessage(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)
	contentLength := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = line[:len(line)-1]

		if len(line) == 0 {
			break
		}

		if len(line) > len(ContentLength) && line[:len(ContentLength)] == ContentLength {
			_, err := fmt.Sscanf(line, "Content-Length: %d", &contentLength)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length header: %w", err)
			}
		}
	}

	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return body, nil
}

func writeMessage(w io.Writer, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func serve(logf LogFunc, handler *Handler) {
	for {
		body, err := readMessage(os.Stdin)
		if err != nil {
			if err == io.EOF {
				return
			}
			logf("read error: %v", err)
			return
		}

		var req jsonRPCMessage
		if err := json.Unmarshal(body, &req); err != nil {
			logf("failed to unmarshal request: %v", err)
			continue
		}

		if req.Method == "" {
			continue
		}

		logf("--> %s", req.Method)

		resp := handler.Handle(req)

		if resp != nil {
			logf("<-- %s (id=%v)", req.Method, req.ID)
			if err := writeMessage(os.Stdout, resp); err != nil {
				logf("write error: %v", err)
				return
			}
		}
	}
}
