package master

import (
	"context"
	"strings"
	"testing"

	"github.com/centrifugal/centrifuge"
	"github.com/gogo/protobuf/proto"
	wgproto "github.com/pnforge/wgadmin/wglib/proto"
)

func TestIncomingMessage(t *testing.T) {
	m := Master{
		logger: dummyLogger,
	}

	sink.Reset()

	var msg centrifuge.Raw

	ctx := context.WithValue(context.Background(), keyCtxRemoteAddr, "127.0.0.1") // nolint:golint

	m.processIncomingMessage(ctx, msg)

	out := sink.String()
	if !strings.Contains(out, "ignoring message") {
		t.Fatalf("message without id should be ignored")
	}

	sink.Reset()

	var packet wgproto.WGPacket

	packet.UUID = "123123"
	packet.PacketType = wgproto.PacketType_PT_RESULT

	bts, _ := proto.Marshal(&packet)
	msg = bts

	m.processIncomingMessage(ctx, msg)

	out = sink.String()
	if !strings.Contains(out, "task not found") {
		t.Fatalf("task shouldn't be found. Logs:\n%s\n", out)
	}

	sink.Reset()

	packetChan := make(chan wgproto.WGPacket, 1)
	m.tasks.Store(packet.UUID, packetChan)
	m.processIncomingMessage(ctx, msg)

	if len(packetChan) != 1 {
		t.Fatalf("expected chan to be of size 1")
	}

	packetCopy := <-packetChan
	if packetCopy != packet {
		t.Fatalf("expected packets to be equal")
	}
}
