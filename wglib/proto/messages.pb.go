// Code generated by protoc-gen-go.
// source: messages.proto
// DO NOT EDIT!

/*
Package wgproto is a generated protocol buffer package.

It is generated from these files:
	messages.proto

It has these top-level messages:
	WGPacket
	Result
	Interface
	Interfaces
	SyncResponse
	RemoveRequest
*/
package wgproto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PacketType int32

const (
	PacketType_PT_ERROR         PacketType = 0
	PacketType_PT_RESULT        PacketType = 1
	PacketType_PT_IFS_REQUEST   PacketType = 2
	PacketType_PT_IFS_RESPONSE  PacketType = 3
	PacketType_PT_SYNC_REQUEST  PacketType = 4
	PacketType_PT_SYNC_RESPONSE PacketType = 5
	PacketType_PT_ADD_IF        PacketType = 6
	PacketType_PT_REMOVE_IF     PacketType = 7
)

var PacketType_name = map[int32]string{
	0: "PT_ERROR",
	1: "PT_RESULT",
	2: "PT_IFS_REQUEST",
	3: "PT_IFS_RESPONSE",
	4: "PT_SYNC_REQUEST",
	5: "PT_SYNC_RESPONSE",
	6: "PT_ADD_IF",
	7: "PT_REMOVE_IF",
}
var PacketType_value = map[string]int32{
	"PT_ERROR":         0,
	"PT_RESULT":        1,
	"PT_IFS_REQUEST":   2,
	"PT_IFS_RESPONSE":  3,
	"PT_SYNC_REQUEST":  4,
	"PT_SYNC_RESPONSE": 5,
	"PT_ADD_IF":        6,
	"PT_REMOVE_IF":     7,
}

func (x PacketType) String() string {
	return proto.EnumName(PacketType_name, int32(x))
}
func (PacketType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type WGPacket struct {
	PacketType PacketType           `protobuf:"varint,1,opt,name=PacketType,json=packetType,enum=wgproto.PacketType" json:"PacketType,omitempty"`
	UUID       string               `protobuf:"bytes,2,opt,name=UUID,json=uUID" json:"UUID,omitempty"`
	Error      string               `protobuf:"bytes,3,opt,name=Error,json=error" json:"Error,omitempty"`
	Payload    *google_protobuf.Any `protobuf:"bytes,4,opt,name=Payload,json=payload" json:"Payload,omitempty"`
}

func (m *WGPacket) Reset()                    { *m = WGPacket{} }
func (m *WGPacket) String() string            { return proto.CompactTextString(m) }
func (*WGPacket) ProtoMessage()               {}
func (*WGPacket) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *WGPacket) GetPacketType() PacketType {
	if m != nil {
		return m.PacketType
	}
	return PacketType_PT_ERROR
}

func (m *WGPacket) GetUUID() string {
	if m != nil {
		return m.UUID
	}
	return ""
}

func (m *WGPacket) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *WGPacket) GetPayload() *google_protobuf.Any {
	if m != nil {
		return m.Payload
	}
	return nil
}

type Result struct {
	Success bool   `protobuf:"varint,1,opt,name=Success,json=success" json:"Success,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=Error,json=error" json:"Error,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Result) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *Result) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

type Interface struct {
	PubKey string `protobuf:"bytes,1,opt,name=PubKey,json=pubKey" json:"PubKey,omitempty"`
	IP     string `protobuf:"bytes,2,opt,name=IP,json=iP" json:"IP,omitempty"`
}

func (m *Interface) Reset()                    { *m = Interface{} }
func (m *Interface) String() string            { return proto.CompactTextString(m) }
func (*Interface) ProtoMessage()               {}
func (*Interface) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Interface) GetPubKey() string {
	if m != nil {
		return m.PubKey
	}
	return ""
}

func (m *Interface) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

type Interfaces struct {
	Interfaces []*Interface `protobuf:"bytes,1,rep,name=Interfaces,json=interfaces" json:"Interfaces,omitempty"`
}

func (m *Interfaces) Reset()                    { *m = Interfaces{} }
func (m *Interfaces) String() string            { return proto.CompactTextString(m) }
func (*Interfaces) ProtoMessage()               {}
func (*Interfaces) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Interfaces) GetInterfaces() []*Interface {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

type SyncResponse struct {
	Interfaces []*Interface `protobuf:"bytes,1,rep,name=Interfaces,json=interfaces" json:"Interfaces,omitempty"`
}

func (m *SyncResponse) Reset()                    { *m = SyncResponse{} }
func (m *SyncResponse) String() string            { return proto.CompactTextString(m) }
func (*SyncResponse) ProtoMessage()               {}
func (*SyncResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *SyncResponse) GetInterfaces() []*Interface {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

type RemoveRequest struct {
	Keys []string `protobuf:"bytes,1,rep,name=Keys,json=keys" json:"Keys,omitempty"`
}

func (m *RemoveRequest) Reset()                    { *m = RemoveRequest{} }
func (m *RemoveRequest) String() string            { return proto.CompactTextString(m) }
func (*RemoveRequest) ProtoMessage()               {}
func (*RemoveRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RemoveRequest) GetKeys() []string {
	if m != nil {
		return m.Keys
	}
	return nil
}

func init() {
	proto.RegisterType((*WGPacket)(nil), "wgproto.WGPacket")
	proto.RegisterType((*Result)(nil), "wgproto.Result")
	proto.RegisterType((*Interface)(nil), "wgproto.Interface")
	proto.RegisterType((*Interfaces)(nil), "wgproto.Interfaces")
	proto.RegisterType((*SyncResponse)(nil), "wgproto.SyncResponse")
	proto.RegisterType((*RemoveRequest)(nil), "wgproto.RemoveRequest")
	proto.RegisterEnum("wgproto.PacketType", PacketType_name, PacketType_value)
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 411 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x51, 0x41, 0x8f, 0x93, 0x50,
	0x10, 0x16, 0x4a, 0xa1, 0xcc, 0x76, 0x2b, 0x99, 0x6d, 0x0c, 0x7a, 0x6a, 0xf0, 0xd2, 0x78, 0x60,
	0x93, 0xf6, 0xe2, 0xd1, 0xd5, 0xb2, 0x86, 0xac, 0x6e, 0x9f, 0x0f, 0xd0, 0x78, 0x22, 0x14, 0x67,
	0xc9, 0x66, 0xbb, 0x80, 0x3c, 0xd0, 0xf0, 0x53, 0xbc, 0xf8, 0x5b, 0x4d, 0x1f, 0x48, 0x7b, 0xde,
	0xdb, 0x7c, 0xdf, 0x7c, 0xdf, 0x7c, 0xf3, 0xe6, 0xc1, 0xec, 0x91, 0x84, 0x48, 0x32, 0x12, 0x6e,
	0x59, 0x15, 0x75, 0x81, 0xc6, 0xef, 0x4c, 0x16, 0xaf, 0x5e, 0x66, 0x45, 0x91, 0xed, 0xe9, 0x52,
	0xa2, 0x5d, 0x73, 0x77, 0x99, 0xe4, 0x6d, 0xa7, 0x71, 0xfe, 0x28, 0x30, 0xf9, 0xf6, 0x91, 0x25,
	0xe9, 0x03, 0xd5, 0xb8, 0x06, 0xe8, 0xaa, 0xb0, 0x2d, 0xc9, 0x56, 0x16, 0xca, 0x72, 0xb6, 0xba,
	0x70, 0xfb, 0x29, 0xee, 0xb1, 0xc5, 0xa1, 0x1c, 0x6a, 0x44, 0xd0, 0xa2, 0xc8, 0xdf, 0xd8, 0xea,
	0x42, 0x59, 0x9a, 0x5c, 0x6b, 0x22, 0x7f, 0x83, 0x73, 0x18, 0x7b, 0x55, 0x55, 0x54, 0xf6, 0x48,
	0x92, 0x63, 0x3a, 0x00, 0x74, 0xc1, 0x60, 0x49, 0xbb, 0x2f, 0x92, 0x1f, 0xb6, 0xb6, 0x50, 0x96,
	0x67, 0xab, 0xb9, 0xdb, 0x2d, 0xe6, 0xfe, 0x5f, 0xcc, 0xbd, 0xca, 0x5b, 0x6e, 0x94, 0x9d, 0xc8,
	0x79, 0x0b, 0x3a, 0x27, 0xd1, 0xec, 0x6b, 0xb4, 0xc1, 0x08, 0x9a, 0x34, 0x25, 0x21, 0xe4, 0x56,
	0x13, 0x6e, 0x88, 0x0e, 0x1e, 0x93, 0xd4, 0x93, 0x24, 0x67, 0x0d, 0xa6, 0x9f, 0xd7, 0x54, 0xdd,
	0x25, 0x29, 0xe1, 0x0b, 0xd0, 0x59, 0xb3, 0xbb, 0xa1, 0x56, 0x7a, 0x4d, 0xae, 0x97, 0x12, 0xe1,
	0x0c, 0x54, 0x9f, 0xf5, 0x3e, 0xf5, 0x9e, 0x39, 0xef, 0x00, 0x06, 0x93, 0xc0, 0xd5, 0x29, 0xb2,
	0x95, 0xc5, 0x68, 0x79, 0xb6, 0xc2, 0xe1, 0x16, 0x43, 0x8b, 0xc3, 0xfd, 0xa0, 0x72, 0xde, 0xc3,
	0x34, 0x68, 0xf3, 0x94, 0x93, 0x28, 0x8b, 0x5c, 0xd0, 0x93, 0x66, 0xbc, 0x86, 0x73, 0x4e, 0x8f,
	0xc5, 0x2f, 0xe2, 0xf4, 0xb3, 0x21, 0x51, 0x1f, 0xee, 0x7b, 0x43, 0x6d, 0x67, 0x37, 0xb9, 0xf6,
	0x40, 0xad, 0x78, 0xf3, 0x57, 0x39, 0xfd, 0x29, 0x9c, 0xc2, 0x84, 0x85, 0xb1, 0xc7, 0xf9, 0x96,
	0x5b, 0xcf, 0xf0, 0x1c, 0x4c, 0x16, 0xc6, 0xdc, 0x0b, 0xa2, 0x4f, 0xa1, 0xa5, 0x20, 0xc2, 0x8c,
	0x85, 0xb1, 0x7f, 0x1d, 0xc4, 0xdc, 0xfb, 0x12, 0x79, 0x41, 0x68, 0xa9, 0x78, 0x01, 0xcf, 0x07,
	0x2e, 0x60, 0xdb, 0xdb, 0xc0, 0xb3, 0x46, 0x3d, 0x19, 0x7c, 0xbf, 0xfd, 0x30, 0x28, 0x35, 0x9c,
	0x83, 0x75, 0x24, 0x7b, 0xe9, 0xb8, 0x8f, 0xb8, 0xda, 0x6c, 0x62, 0xff, 0xda, 0xd2, 0xd1, 0x82,
	0xa9, 0x4c, 0xfc, 0xbc, 0xfd, 0xea, 0x1d, 0x18, 0x63, 0xa7, 0xcb, 0x27, 0xae, 0xff, 0x05, 0x00,
	0x00, 0xff, 0xff, 0xb6, 0x6c, 0xff, 0xe4, 0x93, 0x02, 0x00, 0x00,
}
