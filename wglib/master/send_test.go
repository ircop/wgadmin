package master

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	wgproto "github.com/pnforge/wgadmin/wglib/proto"
	"github.com/stretchr/testify/require"
)

func TestSendPacket(t *testing.T) {
	m := Master{
		logger: dummyLogger,
	}

	var fs fakeSender
	err := m.sendPacket("test-id", wgproto.PacketType_PT_IFS_REQUEST, &fs)

	require.Nil(t, err, "unexpected error in send request: %w", err)
	require.NotEqual(t, 0, len(fs.data), "empty data in sender")

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)

	require.Nil(t, err, "failed to unmarshal sender data")
	require.Equal(t, "test-id", packet.UUID, "wrong packet id in sender data")
	require.Equal(t, wgproto.PacketType_PT_IFS_REQUEST, packet.PacketType, "wrong packet type in sender data")
}

func TestPutRequest(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	interfaces := []Interface{{
		Key: "testkey",
		IP:  "1.2.3.4",
	}}

	err := m.sendPutRequest("test-id", interfaces, &fs, false)

	require.Nil(t, err, "unexpected error in put request: %w", err)
	require.NotEqual(t, 0, len(fs.data), "empty data in sender")

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)

	require.Nil(t, err, "failed to unmarshal sender data")
	require.Equal(t, "test-id", packet.UUID, "wrong packet id in sender data")
	require.Equal(t, wgproto.PacketType_PT_ADD_IF, packet.PacketType, "wrong packet type in sender data")

	var p2 wgproto.Interfaces
	err = proto.Unmarshal(packet.Payload.Value, &p2)
	require.Nil(t, err, "failed to unmarshal underlying packet")
	require.Equal(t, 1, len(p2.Interfaces), "wrong interfaces list len")
	require.Equal(t, interfaces[0].Key, p2.Interfaces[0].PubKey, "keys should be equal")
	require.Equal(t, interfaces[0].IP, p2.Interfaces[0].IP, "ips should be equal")
}

func TestDelRequest(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	err := m.sendDelRequest("test-id", []string{"test-key"}, &fs)
	require.Nil(t, err, "error sending del-request")

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)
	require.Nil(t, err, "failed to unmarshal del-request")
	require.Equal(t, "test-id", packet.UUID, "packet id does not match")
	require.Equal(t, wgproto.PacketType_PT_REMOVE_IF, packet.PacketType, "packet type does not match")

	var p2 wgproto.RemoveRequest
	err = proto.Unmarshal(packet.Payload.Value, &p2)
	require.Nil(t, err, "failed to unmarshal underlying packet")
	require.Equal(t, "test-key", p2.Keys[0], "underlying packet pub-key does not match")
}
