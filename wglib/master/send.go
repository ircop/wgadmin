package master

import (
	"errors"
	"fmt"
	"net"

	"github.com/centrifugal/centrifuge"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	wgproto "github.com/pnforge/wgadmin/wglib/proto"
)

type PutRequest struct {
	Keys []string `json:"keys"`
	IP   string   `json:"ip"`
}

var ErrMarshal = errors.New("failed to marshal request")
var ErrSendingPacket = errors.New("failed to send packet")
var ErrWrongID = errors.New("wrong packet ID")
var ErrWrongIP = errors.New("wrong ip address")

type Sender interface {
	Send(centrifuge.Raw) error
}

func (m *Master) sendPacket(id string, ptype wgproto.PacketType, client Sender) error {
	if id == "" {
		return ErrWrongID
	}

	var msg wgproto.WGPacket

	msg.PacketType = ptype
	msg.UUID = id

	data, err := proto.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrMarshal, err)
	}

	if err = client.Send(data); err != nil {
		return fmt.Errorf("%s: %w", ErrSendingPacket, err)
	}

	return nil
}

func (m *Master) sendPutRequest(id string, interfaces []Interface, client Sender, syncPacket bool) error {
	if id == "" {
		return ErrWrongID
	}

	var msg wgproto.WGPacket

	msg.UUID = id
	if syncPacket {
		msg.PacketType = wgproto.PacketType_PT_SYNC_RESPONSE
	} else {
		msg.PacketType = wgproto.PacketType_PT_ADD_IF
	}

	var protoIfs wgproto.Interfaces
	protoIfs.Interfaces = make([]*wgproto.Interface, len(interfaces))

	for i := range interfaces {
		if net.ParseIP(interfaces[i].IP) == nil {
			return fmt.Errorf("%w: %s", ErrWrongIP, interfaces[i].IP)
		}

		protoIfs.Interfaces[i] = &wgproto.Interface{
			PubKey: interfaces[i].Key,
			IP:     interfaces[i].IP,
		}
	}

	data, err := proto.Marshal(&protoIfs)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrMarshal, err)
	}

	msg.Payload = &any.Any{TypeUrl: wgproto.PacketType_PT_ADD_IF.String(), Value: data}

	data, err = proto.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrMarshal, err)
	}

	if err = client.Send(data); err != nil {
		return fmt.Errorf("%s: %w", ErrSendingPacket, err)
	}

	return nil
}

func (m *Master) sendDelRequest(id string, keys []string, client Sender) error {
	if id == "" || len(keys) == 0 {
		return ErrWrongParams
	}

	var msg wgproto.WGPacket

	var rmReq wgproto.RemoveRequest
	rmReq.Keys = keys

	msg.UUID = id
	msg.PacketType = wgproto.PacketType_PT_REMOVE_IF

	data, err := proto.Marshal(&rmReq)
	if err != nil {
		return ErrMarshal
	}

	msg.Payload = &any.Any{
		TypeUrl: wgproto.PacketType_PT_REMOVE_IF.String(),
		Value:   data,
	}

	data, err = proto.Marshal(&msg)
	if err != nil {
		return ErrMarshal
	}

	if err = client.Send(data); err != nil {
		return fmt.Errorf("%s: %w", ErrSendingPacket, err)
	}

	return nil
}
