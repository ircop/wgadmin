package master

import (
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	wgproto "github.com/ircop/wgadmin/wglib/proto"
	"github.com/stretchr/testify/require"
)

func TestPeers(t *testing.T) {
	m := Master{}

	peers := m.Peers()
	require.Equal(t, 0, len(peers), "expected peers count to be zero")

	m.clients.Store("test", "dummy")
	peers = m.Peers()
	require.Equal(t, 1, len(peers))
}

func TestPeerFails(t *testing.T) {
	var fs fakeSender

	m := Master{logger: dummyLogger}

	peer := "127.0.0.1"

	// 1: test non-existing peer
	ifs, errIfs := m.PeerInterfaces(peer)
	errDel := m.DelInterfacesPeer(peer, []string{"testkey"})
	errPut := m.PutInterfacesPeer(peer, []Interface{{
		Key: "testkey",
		IP:  "1.2.3.4",
	}}, false)

	require.Nil(t, ifs, "interfaces should be nil for non-existing peer")
	require.Equal(t, ErrPeerNodFound, errIfs, "expected PeerNotFound error for non-existing peer")
	require.Equal(t, ErrPeerNodFound, errDel, "expected PeerNotFound error for non-existing peer")
	require.Equal(t, ErrPeerNodFound, errPut, "expected PeerNotFound error for non-existing peer")

	// 2: check sending data with timeout
	m.taskTimeout = time.Millisecond * 100
	m.clients.Store(peer, &fs)

	ifs, errIfs = m.PeerInterfaces(peer)
	errDel = m.DelInterfacesPeer(peer, []string{"testkey"})
	errPut = m.PutInterfacesPeer(peer, []Interface{{
		Key: "testkey",
		IP:  "1.2.3.4",
	}}, false)

	require.Nil(t, ifs, "interfaces should be nil for non-existing peer")
	require.NotNil(t, errIfs)
	require.NotNil(t, errDel)
	require.NotNil(t, errPut)
	require.True(t, strings.Contains(errIfs.Error(), ErrTaskTimeout.Error()), "expected timed-out error")
	require.True(t, strings.Contains(errDel.Error(), ErrTaskTimeout.Error()), "expected timed-out error")
	require.True(t, strings.Contains(errPut.Error(), ErrTaskTimeout.Error()), "expected timed-out error")

	// 3: check that result chan is empty and we haven't leak
	require.Nil(t, firstTaskChan(&m))
}

func TestPeerInterfacesData(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	m.clients.Store("127.0.0.1", &fs)

	m.taskTimeout = time.Millisecond * 500
	fs.data = []byte{}

	go func() {
		time.Sleep(time.Millisecond * 50)
		ch := firstTaskChan(&m)
		require.NotNil(t, ch)

		var response wgproto.WGPacket
		response.PacketType = wgproto.PacketType_PT_IFS_RESPONSE

		var ifsresponse = wgproto.Interfaces{}
		ifsresponse.Interfaces = []*wgproto.Interface{{
			PubKey: "testkey",
			IP:     "1.2.3.4",
		}}

		data, err := proto.Marshal(&ifsresponse)
		require.Nil(t, err)

		response.Payload = &any.Any{
			TypeUrl: wgproto.PacketType_PT_IFS_RESPONSE.String(),
			Value:   data,
		}

		ch <- response
	}()

	ifs, err := m.PeerInterfaces("127.0.0.1")
	require.Nil(t, err)

	// check sent data
	require.NotNil(t, fs.data)
	require.NotEqual(t, 0, len(fs.data))

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)

	require.Nil(t, err, "failed to unmarshal sent data")
	require.Equal(t, wgproto.PacketType_PT_IFS_REQUEST, packet.PacketType, "ifrequest packet type mismatch")

	// check resutls
	require.Nil(t, err)
	require.NotNil(t, ifs)
	require.Equal(t, 1, len(ifs))
	require.Equal(t, "testkey", ifs[0].Key)
	require.Equal(t, "1.2.3.4", ifs[0].IP)
}

func TestPeerDelData(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	m.clients.Store("127.0.0.1", &fs)

	m.taskTimeout = time.Millisecond * 500
	fs.data = []byte{}

	go func() {
		time.Sleep(time.Millisecond * 50)
		ch := firstTaskChan(&m)
		require.NotNil(t, ch)

		var response wgproto.WGPacket
		response.PacketType = wgproto.PacketType_PT_RESULT

		delresponse := wgproto.Result{
			Success: true,
			Error:   "123",
		}

		data, err := proto.Marshal(&delresponse)
		require.Nil(t, err)

		response.Payload = &any.Any{
			TypeUrl: wgproto.PacketType_PT_RESULT.String(),
			Value:   data,
		}

		ch <- response
	}()

	err := m.DelInterfacesPeer("127.0.0.1", []string{"testkey"})
	require.Nil(t, err)

	// check sent data
	require.NotNil(t, fs.data)
	require.NotEqual(t, 0, len(fs.data))

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)

	require.Nil(t, err, "failed to unmarshal sent data")
	require.Equal(t, wgproto.PacketType_PT_REMOVE_IF, packet.PacketType, "ifrequest packet type mismatch")
}

func TestPeerPutData(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	m.clients.Store("127.0.0.1", &fs)

	m.taskTimeout = time.Millisecond * 500
	fs.data = []byte{}

	go func() {
		time.Sleep(time.Millisecond * 50)
		ch := firstTaskChan(&m)
		require.NotNil(t, ch)

		var response wgproto.WGPacket
		response.PacketType = wgproto.PacketType_PT_RESULT

		putresponse := wgproto.Result{
			Success: true,
			Error:   "123",
		}

		data, err := proto.Marshal(&putresponse)
		require.Nil(t, err)

		response.Payload = &any.Any{
			TypeUrl: wgproto.PacketType_PT_RESULT.String(),
			Value:   data,
		}

		ch <- response
	}()

	err := m.PutInterfacesPeer("127.0.0.1", []Interface{{
		Key: "testkey",
		IP:  "1.2.3.4",
	}}, false)
	require.Nil(t, err)

	// check sent data
	require.NotNil(t, fs.data)
	require.NotEqual(t, 0, len(fs.data))

	var packet wgproto.WGPacket
	err = proto.Unmarshal(fs.data, &packet)

	require.Nil(t, err, "failed to unmarshal sent data")
	require.Equal(t, wgproto.PacketType_PT_ADD_IF, packet.PacketType, "ifrequest packet type mismatch")
}

func TestAllPeersFuncs(t *testing.T) {
	m := Master{logger: dummyLogger}

	var fs fakeSender

	m.clients.Store("127.0.0.1", &fs)
	m.clients.Store("127.0.0.2", &fs)
	m.taskTimeout = time.Millisecond * 100

	errs := m.DelInterfacesAllPeers([]string{"1", "2"})
	require.Equal(t, 2, len(errs))

	errs = m.PutInterfacesAllPeers([]Interface{{
		Key: "testkey",
		IP:  "1.2.3.4",
	}}, false)
	require.Equal(t, 2, len(errs))
}

func firstTaskChan(m *Master) chan wgproto.WGPacket {
	var ch chan wgproto.WGPacket = nil

	m.tasks.Range(func(k, v interface{}) bool {
		ch = v.(chan wgproto.WGPacket)
		return false
	})

	return ch
}
