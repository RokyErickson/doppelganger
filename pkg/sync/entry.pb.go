// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sync/entry.proto

package sync // import "github.com/RokyErickson/doppelganger/pkg/sync"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const _ = proto.ProtoPackageIsVersion2

type EntryKind int32

const (
	EntryKind_Directory EntryKind = 0
	EntryKind_File      EntryKind = 1
	EntryKind_Symlink   EntryKind = 2
)

var EntryKind_name = map[int32]string{
	0: "Directory",
	1: "File",
	2: "Symlink",
}
var EntryKind_value = map[string]int32{
	"Directory": 0,
	"File":      1,
	"Symlink":   2,
}

func (x EntryKind) String() string {
	return proto.EnumName(EntryKind_name, int32(x))
}
func (EntryKind) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_entry_29deca7dbfaed434, []int{0}
}

type Entry struct {
	Kind                 EntryKind         `protobuf:"varint,1,opt,name=kind,proto3,enum=sync.EntryKind" json:"kind,omitempty"`
	Contents             map[string]*Entry `protobuf:"bytes,5,rep,name=contents,proto3" json:"contents,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Digest               []byte            `protobuf:"bytes,8,opt,name=digest,proto3" json:"digest,omitempty"`
	Executable           bool              `protobuf:"varint,9,opt,name=executable,proto3" json:"executable,omitempty"`
	Target               string            `protobuf:"bytes,12,opt,name=target,proto3" json:"target,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Entry) Reset()         { *m = Entry{} }
func (m *Entry) String() string { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()    {}
func (*Entry) Descriptor() ([]byte, []int) {
	return fileDescriptor_entry_29deca7dbfaed434, []int{0}
}
func (m *Entry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Entry.Unmarshal(m, b)
}
func (m *Entry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Entry.Marshal(b, m, deterministic)
}
func (dst *Entry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Entry.Merge(dst, src)
}
func (m *Entry) XXX_Size() int {
	return xxx_messageInfo_Entry.Size(m)
}
func (m *Entry) XXX_DiscardUnknown() {
	xxx_messageInfo_Entry.DiscardUnknown(m)
}

var xxx_messageInfo_Entry proto.InternalMessageInfo

func (m *Entry) GetKind() EntryKind {
	if m != nil {
		return m.Kind
	}
	return EntryKind_Directory
}

func (m *Entry) GetContents() map[string]*Entry {
	if m != nil {
		return m.Contents
	}
	return nil
}

func (m *Entry) GetDigest() []byte {
	if m != nil {
		return m.Digest
	}
	return nil
}

func (m *Entry) GetExecutable() bool {
	if m != nil {
		return m.Executable
	}
	return false
}

func (m *Entry) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func init() {
	proto.RegisterType((*Entry)(nil), "sync.Entry")
	proto.RegisterMapType((map[string]*Entry)(nil), "sync.Entry.ContentsEntry")
	proto.RegisterEnum("sync.EntryKind", EntryKind_name, EntryKind_value)
}

func init() { proto.RegisterFile("sync/entry.proto", fileDescriptor_entry_29deca7dbfaed434) }

var fileDescriptor_entry_29deca7dbfaed434 = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xdb, 0x4b, 0xc3, 0x30,
	0x18, 0xc5, 0xcd, 0x6e, 0xb6, 0x5f, 0x37, 0x2d, 0x79, 0x90, 0xe8, 0x83, 0xc4, 0x0b, 0x12, 0x04,
	0x5b, 0x9c, 0x08, 0xe2, 0xa3, 0x37, 0x04, 0xdf, 0xe2, 0x9b, 0x6f, 0x5d, 0xf6, 0xd1, 0x85, 0x76,
	0xc9, 0xe8, 0xd2, 0x61, 0xff, 0x78, 0x41, 0x52, 0xc7, 0xa8, 0x6f, 0x39, 0xe7, 0x77, 0xc2, 0x39,
	0x09, 0xc4, 0xeb, 0xc6, 0xa8, 0x14, 0x8d, 0xab, 0x9a, 0x64, 0x55, 0x59, 0x67, 0xe9, 0xc0, 0x3b,
	0xe7, 0x3f, 0x04, 0x86, 0xaf, 0xde, 0xa5, 0x17, 0x30, 0x28, 0xb4, 0x99, 0x33, 0xc2, 0x89, 0x38,
	0x98, 0x1e, 0x26, 0x1e, 0x27, 0x2d, 0xfa, 0xd0, 0x66, 0x2e, 0x5b, 0x48, 0xef, 0x21, 0x50, 0xd6,
	0x38, 0x34, 0x6e, 0xcd, 0x86, 0xbc, 0x2f, 0xa2, 0xe9, 0x71, 0x27, 0x98, 0x3c, 0x6f, 0x59, 0xab,
	0xe4, 0x2e, 0x4a, 0x8f, 0x60, 0x34, 0xd7, 0x39, 0xae, 0x1d, 0x0b, 0x38, 0x11, 0x63, 0xb9, 0x55,
	0xf4, 0x14, 0x00, 0xbf, 0x51, 0xd5, 0x2e, 0x9b, 0x95, 0xc8, 0x42, 0x4e, 0x44, 0x20, 0x3b, 0x8e,
	0xbf, 0xe7, 0xb2, 0x2a, 0x47, 0xc7, 0xc6, 0x9c, 0x88, 0x50, 0x6e, 0xd5, 0xc9, 0x3b, 0x4c, 0xfe,
	0x55, 0xd1, 0x18, 0xfa, 0x05, 0x36, 0xed, 0xf6, 0x50, 0xfa, 0x23, 0x3d, 0x83, 0xe1, 0x26, 0x2b,
	0x6b, 0x64, 0x3d, 0x4e, 0x44, 0x34, 0x8d, 0x3a, 0x33, 0xe5, 0x1f, 0x79, 0xec, 0x3d, 0x90, 0xeb,
	0x5b, 0x08, 0x77, 0x6f, 0xa4, 0x13, 0x08, 0x5f, 0x74, 0x85, 0xca, 0xd9, 0xaa, 0x89, 0xf7, 0x68,
	0x00, 0x83, 0x37, 0x5d, 0x62, 0x4c, 0x68, 0x04, 0xfb, 0x9f, 0xcd, 0xb2, 0xd4, 0xa6, 0x88, 0x7b,
	0x4f, 0x57, 0x5f, 0x97, 0xb9, 0x76, 0x8b, 0x7a, 0x96, 0x28, 0xbb, 0x4c, 0x17, 0xd9, 0xc6, 0xaa,
	0x1b, 0x6d, 0xd3, 0x65, 0xed, 0xb2, 0x1c, 0x4d, 0xba, 0x2a, 0xf2, 0xd4, 0x77, 0xcd, 0x46, 0xed,
	0x3f, 0xdf, 0xfd, 0x06, 0x00, 0x00, 0xff, 0xff, 0x12, 0x4a, 0xdb, 0x26, 0x7b, 0x01, 0x00, 0x00,
}
