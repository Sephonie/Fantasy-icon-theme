// Code generated by protoc-gen-go.
// source: proto3_proto/proto3.proto
// DO NOT EDIT!

/*
Package proto3_proto is a generated protocol buffer package.

It is generated from these files:
	proto3_proto/proto3.proto

It has these top-level messages:
	Message
	Nested
	MessageWithMap
	IntMap
	IntMaps
*/
package proto3_proto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"
import testdata "github.com/golang/protobuf/proto/testdata"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Message_Humour int32

const (
	Message_UNKNOWN     Message_Humour = 0
	Message_PUNS        Message_Humour = 1
	Message_SLAPSTICK   Message_Humour = 2
	Message_BILL_BAILEY Message_Humour = 3
)

var Message_Humour_name = map[int32]string{
	0: "UNKNOWN",
	1: "PUNS",
	2: "SLAPSTICK",
	3: "BILL_BAILEY",
}
var Message_Humour_value = map[string]int32{
	"UNKNOWN":     0,
	"PUNS":        1,
	"SLAPSTICK":   2,
	"BILL_BAILEY": 3,
}

func (x Message_Humour) String() string {
	return proto.EnumName(Message_Humour_name, int32(x))
}
func (Message_Humour) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Message struct {
	Name         string                           `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Hilarity     Message_Humour                   `protobuf:"varint,2,opt,name=hilarity,enum=proto3_proto.Message_Humour" json:"hilarity,omitempty"`
	HeightInCm   uint32                           `protobuf:"varint,3,opt,name=height_in_cm,json=heightInCm" json:"height_in_cm,omitempty"`
	Data         []byte                           `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	ResultCount  int64                            `protobuf:"varint,7,opt,name=result_count,json=resultCount" json:"result_count,omitempty"`
	TrueScotsman bool                             `protobuf:"varint,8,opt,name=true_scotsman,json=trueScotsman" json:"true_scotsman,omitempty"`
	Score        float32                          `protobuf:"fixed32,9,opt,name=score" json:"score,omitempty"`
	Key          []uint64                         `protobuf:"varint,5,rep,packed,name=key" json:"key,omitempty"`
	ShortKey     []int32                          `protobuf:"varint,19,rep,packed,name=short_key,json=shortKey" json:"short_key,omitempty"`
	Nested       *Nested                          `protobuf:"bytes,6,opt,name=nested" json:"nested,omitempty"`
	RFunny       []Message_Humour                 `protobuf:"varint,16,rep,packed,name=r_funny,json=rFunny,enum=proto3_proto.Message_Humour" json:"r_funny,omitempty"`
	Terrain      map[string]*Nested               `protobuf:"bytes,10,rep,name=terrain" json:"terrain,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Proto2Field  *testdata.SubDefaults            `protobuf:"bytes,11,opt,name=proto2_field,json=proto2Field" json:"proto2_field,omitempty"`
	Proto2Value  map[string]*testdata.SubDefaults `protobuf:"bytes,13,rep,name=proto2_value,json=proto2Value" json:"proto2_value,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Anything     *google_protobuf.Any             `protobuf:"bytes,14,opt,name=anything" json:"anything,omitempty"`
	ManyThings   []*google_protobuf.Any           `protobuf:"bytes,15,rep,name=many_things,json=manyThings" json:"many_things,omitempty"`
	Submessage   *Message                         `protobuf:"bytes,17,opt,name=submessage" json:"submessage,omitempty"`
	Children     []*Message                       `protobuf:"bytes,18,rep,name=children" json:"children,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Message) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Message) GetHilarity() Message_Humour {
	if m != nil {
		return m.Hilarity
	}
	return Message_UNKNOWN
}

func (m *Message) GetHeightInCm() uint32 {
	if m != nil {
		return m.HeightInCm
	}
	return 0
}

func (m *Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Message) GetResultCount() int64 {
	if m != nil {
		return m.ResultCount
	}
	return 0
}

func (m *Message) GetTrueScotsman() bool {
	if m != nil {
		return m.TrueScotsman
	}
	return false
}

func (m *Message) GetScore() float32 {
	if m != nil {
		return m.Score
	}
	return 0
}

func (m *Message) GetKey() []uint64 {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *Message) GetShortKey() []int32 {
	if m != nil {
		return m.ShortKey
	}
	return nil
}

func (m *Message) GetNested() *Nested {
	if m != nil {
		return m.Nested
	}
	return nil
}

func (m *Message) GetRFunny() []Message_Humour {
	if m != nil {
		return m.RFunny
	}
	return nil
}

func (m *Message) GetTerrain() map[string]*Nested {
	if m != nil {
		return m.Terrain
	}
	return nil
}

func (m *Message) GetProto2Field() *testdata.SubDefaults {
	if m != nil {
		return m.Proto2Field
	}
	return nil
}

func (m *Message) GetProto2Value() map[string]*testdata.SubDefaults {
	if m != nil {
		return m.Proto2Value
	}
	return nil
}

func (m *Message) GetAnything() *google_protobuf.Any {
	if m != nil {
		return m.Anything
	}
	return nil
}

func (m *Message) GetManyThings() []*google_protobuf.Any {
	if m != nil {
		return m.ManyThings
	}
	return nil
}

func (m *Message) GetSubmessage() *Message {
	if m != nil {
		return m.Submessage
	}
	return nil
}

func (m *Message) GetChildren() []*Message {
	if m != nil {
		return m.Children
	}
	return nil
}

type Nested struct {
	Bunny string `protobuf:"bytes,1,opt,name=bunny" json:"bunny,omitempty"`
	Cute  bool   `protobuf:"varint,2,opt,name=cute" json:"cute,omitempty"`
}

func (m *Nested) Reset()                    { *m = Nested{} }
func (m *Nested) String() string            { return proto.CompactTextString(m) }
func (*Nested) ProtoMessage()               {}
func (*Nested) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Nested) GetBunny() string {
	if m != nil {
		return m.Bunny
	}
	return ""
}

func (m *Nested) GetCute() bool {
	if m != nil {
		return m.Cute
	}
	return false
}

type MessageWithMap struct {
	ByteMapping map[bool][]byte `protobuf:"bytes,1,rep,name=byte_mapping,json=byteMapping" json:"byte_mapping,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *MessageWithMap) Reset()                    { *m = MessageWithMap{} }
func (m *MessageWithMap) String() string            { return proto.CompactTextString(m) }
func (*MessageWithMap) ProtoMessage()               {}
func (*MessageWithMap) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MessageWithMap) GetByteMapping() map[bool][]byte {
	if m != nil {
		return m.ByteMapping
	}
	return nil
}

type IntMap struct {
	Rtt map[int32]int32 `protobuf:"bytes,1,rep,name=rtt" json:"rtt,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *IntMap) Reset()                    { *m = IntMap{} }
func (m *IntMap) String() string            { return proto.CompactTextString(m) }
func (*IntMap) ProtoMessage()               {}
func (*IntMap) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *IntMap) GetRtt() map[int32]int32 {
	if m != nil {
		return m.Rtt
	}
	return nil
}

type IntMaps struct {
	Maps []*IntMap `protobuf:"bytes,1,rep,name=maps" json:"maps,omitempty"`
}

func (m *IntMaps) Reset()                    { *m = IntMaps{} }
func (m *IntMaps) String() string            { return proto.CompactTextString(m) }
func (*IntMaps) ProtoMessage()               {}
func (*IntMaps) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *IntMaps) GetMaps() []*IntMap {
	if m != nil {
		return m.Maps
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "proto3_proto.Message")
	proto.RegisterType((*Nested)(nil), "proto3_proto.Nested")
	proto.RegisterType((*MessageWithMap)(nil), "proto3_proto.MessageWithMap")
	proto.RegisterType((*IntMap)(nil), "proto3_proto.IntMap")
	proto.RegisterType((*IntMaps)(nil), "proto3_proto.IntMaps")
	proto.RegisterEnum("proto3_proto.Message_Humour", Message_Humour_name, Message_Humour_value)
}

func init() { proto.RegisterFile("proto3_proto/proto3.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 733 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x53, 0x6d, 0x6f, 0xf3, 0x34,
	0x14, 0x25, 0x4d, 0x5f, 0xd2, 0x9b, 0x74, 0x0b, 0x5e, 0x91, 0xbc, 0x02, 0x52, 0x28, 0x12, 0x8a,
	0x78, 0x49, 0xa1, 0xd3, 0xd0, 0x84, 0x10, 0x68, 0x1b, 0x9b, 0xa8, 0xd6, 0x95, 0xca, 0xdd, 0x98,
	0xf8, 0x14, 0xa5, 0xad, 0xdb, 0x46, 0x34, 0x4e, 0x49, 0x1c, 0xa4, 0xfc, 0x1d, 0xfe, 0x28, 0x8f,
	0x6c, 0xa7, 0x5d, 0x36, 0x65, 0xcf, 0xf3, 0x29, 0xf6, 0xf1, 0xb9, 0xf7, 0x9c, 0x1c, 0x5f, 0xc3,
	0xe9, 0x2e, 0x89, 0x79, 0x7c, 0xe6, 0xcb, 0xcf, 0x40, 0x6d, 0x3c, 0xf9, 0x41, 0x56, 0xf9, 0xa8,
	0x77, 0xba, 0x8e, 0xe3, 0xf5, 0x96, 0x2a, 0xca, 0x3c, 0x5b, 0x0d, 0x02, 0x96, 0x2b, 0x62, 0xef,
	0x84, 0xd3, 0x94, 0x2f, 0x03, 0x1e, 0x0c, 0xc4, 0x42, 0x81, 0xfd, 0xff, 0x5b, 0xd0, 0xba, 0xa7,
	0x69, 0x1a, 0xac, 0x29, 0x42, 0x50, 0x67, 0x41, 0x44, 0xb1, 0xe6, 0x68, 0x6e, 0x9b, 0xc8, 0x35,
	0xba, 0x00, 0x63, 0x13, 0x6e, 0x83, 0x24, 0xe4, 0x39, 0xae, 0x39, 0x9a, 0x7b, 0x34, 0xfc, 0xcc,
	0x2b, 0x0b, 0x7a, 0x45, 0xb1, 0xf7, 0x7b, 0x16, 0xc5, 0x59, 0x42, 0x0e, 0x6c, 0xe4, 0x80, 0xb5,
	0xa1, 0xe1, 0x7a, 0xc3, 0xfd, 0x90, 0xf9, 0x8b, 0x08, 0xeb, 0x8e, 0xe6, 0x76, 0x08, 0x28, 0x6c,
	0xc4, 0xae, 0x23, 0xa1, 0x27, 0xec, 0xe0, 0xba, 0xa3, 0xb9, 0x16, 0x91, 0x6b, 0xf4, 0x05, 0x58,
	0x09, 0x4d, 0xb3, 0x2d, 0xf7, 0x17, 0x71, 0xc6, 0x38, 0x6e, 0x39, 0x9a, 0xab, 0x13, 0x53, 0x61,
	0xd7, 0x02, 0x42, 0x5f, 0x42, 0x87, 0x27, 0x19, 0xf5, 0xd3, 0x45, 0xcc, 0xd3, 0x28, 0x60, 0xd8,
	0x70, 0x34, 0xd7, 0x20, 0x96, 0x00, 0x67, 0x05, 0x86, 0xba, 0xd0, 0x48, 0x17, 0x71, 0x42, 0x71,
	0xdb, 0xd1, 0xdc, 0x1a, 0x51, 0x1b, 0x64, 0x83, 0xfe, 0x37, 0xcd, 0x71, 0xc3, 0xd1, 0xdd, 0x3a,
	0x11, 0x4b, 0xf4, 0x29, 0xb4, 0xd3, 0x4d, 0x9c, 0x70, 0x5f, 0xe0, 0x27, 0x8e, 0xee, 0x36, 0x88,
	0x21, 0x81, 0x3b, 0x9a, 0xa3, 0x6f, 0xa1, 0xc9, 0x68, 0xca, 0xe9, 0x12, 0x37, 0x1d, 0xcd, 0x35,
	0x87, 0xdd, 0x97, 0xbf, 0x3e, 0x91, 0x67, 0xa4, 0xe0, 0xa0, 0x73, 0x68, 0x25, 0xfe, 0x2a, 0x63,
	0x2c, 0xc7, 0xb6, 0xa3, 0x7f, 0x30, 0xa9, 0x66, 0x72, 0x2b, 0xb8, 0xe8, 0x67, 0x68, 0x71, 0x9a,
	0x24, 0x41, 0xc8, 0x30, 0x38, 0xba, 0x6b, 0x0e, 0xfb, 0xd5, 0x65, 0x0f, 0x8a, 0x74, 0xc3, 0x78,
	0x92, 0x93, 0x7d, 0x09, 0xba, 0x00, 0x75, 0xff, 0x43, 0x7f, 0x15, 0xd2, 0xed, 0x12, 0x9b, 0xd2,
	0xe8, 0x27, 0xde, 0xfe, 0xae, 0xbd, 0x59, 0x36, 0xff, 0x8d, 0xae, 0x82, 0x6c, 0xcb, 0x53, 0x62,
	0x2a, 0xea, 0xad, 0x60, 0xa2, 0xd1, 0xa1, 0xf2, 0xdf, 0x60, 0x9b, 0x51, 0xdc, 0x91, 0xe2, 0x5f,
	0x55, 0x8b, 0x4f, 0x25, 0xf3, 0x4f, 0x41, 0x54, 0x06, 0x8a, 0x56, 0x12, 0x41, 0xdf, 0x83, 0x11,
	0xb0, 0x9c, 0x6f, 0x42, 0xb6, 0xc6, 0x47, 0x45, 0x52, 0x6a, 0x0e, 0xbd, 0xfd, 0x1c, 0x7a, 0x97,
	0x2c, 0x27, 0x07, 0x16, 0x3a, 0x07, 0x33, 0x0a, 0x58, 0xee, 0xcb, 0x5d, 0x8a, 0x8f, 0xa5, 0x76,
	0x75, 0x11, 0x08, 0xe2, 0x83, 0xe4, 0xa1, 0x73, 0x80, 0x34, 0x9b, 0x47, 0xca, 0x14, 0xfe, 0xb8,
	0xf8, 0xd7, 0x2a, 0xc7, 0xa4, 0x44, 0x44, 0x3f, 0x80, 0xb1, 0xd8, 0x84, 0xdb, 0x65, 0x42, 0x19,
	0x46, 0x52, 0xea, 0x8d, 0xa2, 0x03, 0xad, 0x37, 0x05, 0xab, 0x1c, 0xf8, 0x7e, 0x72, 0xd4, 0xd3,
	0x90, 0x93, 0xf3, 0x35, 0x34, 0x54, 0x70, 0xb5, 0xf7, 0xcc, 0x86, 0xa2, 0xfc, 0x54, 0xbb, 0xd0,
	0x7a, 0x8f, 0x60, 0xbf, 0x4e, 0xb1, 0xa2, 0xeb, 0x37, 0x2f, 0xbb, 0xbe, 0x71, 0x91, 0xcf, 0x6d,
	0xfb, 0xbf, 0x42, 0x53, 0x0d, 0x14, 0x32, 0xa1, 0xf5, 0x38, 0xb9, 0x9b, 0xfc, 0xf1, 0x34, 0xb1,
	0x3f, 0x42, 0x06, 0xd4, 0xa7, 0x8f, 0x93, 0x99, 0xad, 0xa1, 0x0e, 0xb4, 0x67, 0xe3, 0xcb, 0xe9,
	0xec, 0x61, 0x74, 0x7d, 0x67, 0xd7, 0xd0, 0x31, 0x98, 0x57, 0xa3, 0xf1, 0xd8, 0xbf, 0xba, 0x1c,
	0x8d, 0x6f, 0xfe, 0xb2, 0xf5, 0xfe, 0x10, 0x9a, 0xca, 0xac, 0x78, 0x33, 0x73, 0x39, 0xbe, 0xca,
	0x8f, 0xda, 0x88, 0x57, 0xba, 0xc8, 0xb8, 0x32, 0x64, 0x10, 0xb9, 0xee, 0xff, 0xa7, 0xc1, 0x51,
	0x91, 0xd9, 0x53, 0xc8, 0x37, 0xf7, 0xc1, 0x0e, 0x4d, 0xc1, 0x9a, 0xe7, 0x9c, 0xfa, 0x51, 0xb0