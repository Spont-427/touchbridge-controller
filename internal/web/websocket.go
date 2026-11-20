package web

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/Spont-427/touchbridge-controller/internal/protocol"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

var errWSProtocol = errors.New("websocket protocol error")

type wsConn struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

func headerHasToken(header http.Header, name, expected string) bool {
	for _, value := range header.Values(name) {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), expected) {
				return true
			}
		}
	}
	return false
}

func upgrade(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	if !headerHasToken(r.Header, "Upgrade", "websocket") ||
		!headerHasToken(r.Header, "Connection", "upgrade") ||
		r.Header.Get("Sec-WebSocket-Version") != "13" {
		return nil, errWSProtocol
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	decodedKey, keyErr := base64.StdEncoding.DecodeString(key)
	if keyErr != nil || len(decodedKey) != 16 {
		return nil, errWSProtocol
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("HTTP hijacking unavailable")
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}
	accept := sha1.Sum([]byte(key + wsGUID))
	_, err = fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", base64.StdEncoding.EncodeToString(accept[:]))
	if err == nil {
		err = rw.Flush()
	}
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &wsConn{conn: conn, r: rw.Reader, w: rw.Writer}, nil
}

func (c *wsConn) readMessage() (byte, []byte, error) {
	var h [2]byte
	if _, err := io.ReadFull(c.r, h[:]); err != nil {
		return 0, nil, err
	}
	fin := h[0]&0x80 != 0
	opcode := h[0] & 0x0f
	masked := h[1]&0x80 != 0
	// This server intentionally supports only complete text messages and the
	// three WebSocket control opcodes. Fragmentation/extensions are not needed
	// by the controller and expand the parser's attack surface.
	if h[0]&0x70 != 0 || !fin || !masked ||
		(opcode != 1 && opcode != 8 && opcode != 9 && opcode != 10) {
		return 0, nil, errWSProtocol
	}
	length := uint64(h[1] & 0x7f)
	if length == 126 {
		var ext [2]byte
		if _, err := io.ReadFull(c.r, ext[:]); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
		if length < 126 { // Reject non-canonical length encodings.
			return 0, nil, errWSProtocol
		}
	} else if length == 127 {
		var ext [8]byte
		if _, err := io.ReadFull(c.r, ext[:]); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext[:])
		if length <= 65535 || length&(uint64(1)<<63) != 0 {
			return 0, nil, errWSProtocol
		}
	}
	if length > protocol.MaxMessageBytes || (opcode >= 8 && length > 125) {
		return 0, nil, errWSProtocol
	}
	var mask [4]byte
	if _, err := io.ReadFull(c.r, mask[:]); err != nil {
		return 0, nil, err
	}
	payload := make([]byte, int(length))
	if _, err := io.ReadFull(c.r, payload); err != nil {
		return 0, nil, err
	}
	for i := range payload {
		payload[i] ^= mask[i&3]
	}
	return opcode, payload, nil
}
func (c *wsConn) writeFrame(opcode byte, payload []byte) error {
	if len(payload) > 65536 {
		return errWSProtocol
	}
	if err := c.w.WriteByte(0x80 | opcode); err != nil {
		return err
	}
	switch n := len(payload); {
	case n < 126:
		if err := c.w.WriteByte(byte(n)); err != nil {
			return err
		}
	case n <= 65535:
		if err := c.w.WriteByte(126); err != nil {
			return err
		}
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], uint16(n))
		if _, err := c.w.Write(b[:]); err != nil {
			return err
		}
	default:
		if err := c.w.WriteByte(127); err != nil {
			return err
		}
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(n))
		if _, err := c.w.Write(b[:]); err != nil {
			return err
		}
	}
	if _, err := c.w.Write(payload); err != nil {
		return err
	}
	return c.w.Flush()
}
func (c *wsConn) close() error { _ = c.writeFrame(8, []byte{}); return c.conn.Close() }
