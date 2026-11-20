package discovery

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryProbeGetsServerResponse(t *testing.T) {
	reserved, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	port := reserved.LocalAddr().(*net.UDPAddr).Port
	_ = reserved.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := Start(ctx, port, 8765, "0123456789abcdef0123456789abcdef", "https"); err != nil {
		t.Fatal(err)
	}

	client, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	probe := []byte(Probe)
	if _, err := client.WriteToUDP(probe, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port}); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 256)
	n, _, err := client.ReadFromUDP(buf)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(buf[:n]), "\n")
	if len(lines) != 5 || lines[0] != "TOUCHBRIDGE_SERVER_V1" || lines[1] != strconv.Itoa(8765) || lines[3] != "0123456789abcdef0123456789abcdef" || lines[4] != "https" {
		t.Fatalf("unexpected discovery response %q", string(buf[:n]))
	}
}
