// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sync/problem.proto

package sync // import "github.com/RokyErickson/doppelganger/pkg/sync"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const _ = proto.ProtoPackageIsVersion2

type Problem struct {
	Path                 string   `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Problem) Reset()         { *m = Problem{} }
func (m *Problem) String() string { return proto.CompactTextString(m) }
func (*Problem) ProtoMessage()    {}
func (*Problem) Descriptor() ([]byte, []int) {
	return fileDescriptor_problem_3c4fe5e23768b357, []int{0}
}
func (m *Problem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Problem.Unmarshal(m, b)
}
func (m *Problem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Problem.Marshal(b, m, deterministic)
}
func (dst *Problem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Problem.Merge(dst, src)
}
func (m *Problem) XXX_Size() int {
	return xxx_messageInfo_Problem.Size(m)
}
func (m *Problem) XXX_DiscardUnknown() {
	xxx_messageInfo_Problem.DiscardUnknown(m)
}

var xxx_messageInfo_Problem proto.InternalMessageInfo

func (m *Problem) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Problem) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*Problem)(nil), "sync.Problem")
}

func init() { proto.RegisterFile("sync/problem.proto", fileDescriptor_problem_3c4fe5e23768b357) }

var fileDescriptor_problem_3c4fe5e23768b357 = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2a, 0xae, 0xcc, 0x4b,
	0xd6, 0x2f, 0x28, 0xca, 0x4f, 0xca, 0x49, 0xcd, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62,
	0x01, 0x89, 0x29, 0x19, 0x73, 0xb1, 0x07, 0x40, 0x84, 0x85, 0x84, 0xb8, 0x58, 0x0a, 0x12, 0x4b,
	0x32, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xc0, 0x6c, 0x21, 0x11, 0x2e, 0xd6, 0xd4, 0xa2,
	0xa2, 0xfc, 0x22, 0x09, 0x26, 0xb0, 0x20, 0x84, 0xe3, 0xa4, 0x16, 0xa5, 0x92, 0x9e, 0x59, 0x92,
	0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0xab, 0x9f, 0x91, 0x58, 0x96, 0x9f, 0xac, 0x9b, 0x99, 0xaf,
	0x9f, 0x5b, 0x5a, 0x92, 0x98, 0x9e, 0x9a, 0xa7, 0x5f, 0x90, 0x9d, 0xae, 0x0f, 0x32, 0x3c, 0x89,
	0x0d, 0x6c, 0x93, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x8f, 0xcf, 0x3d, 0x80, 0x7f, 0x00, 0x00,
	0x00,
}