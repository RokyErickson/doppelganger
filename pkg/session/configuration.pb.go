// Code generated by protoc-gen-go. DO NOT EDIT.
// source: session/configuration.proto

package session // import "github.com/RokyErickson/doppelganger/pkg/session"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import filesystem "github.com/RokyErickson/doppelganger/pkg/filesystem"
import sync "github.com/RokyErickson/doppelganger/pkg/sync"

var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const _ = proto.ProtoPackageIsVersion2

type Configuration struct {
	SynchronizationMode    sync.SynchronizationMode `protobuf:"varint,11,opt,name=synchronizationMode,proto3,enum=sync.SynchronizationMode" json:"synchronizationMode,omitempty"`
	MaximumEntryCount      uint64                   `protobuf:"varint,12,opt,name=maximumEntryCount,proto3" json:"maximumEntryCount,omitempty"`
	MaximumStagingFileSize uint64                   `protobuf:"varint,13,opt,name=maximumStagingFileSize,proto3" json:"maximumStagingFileSize,omitempty"`
	SymlinkMode            sync.SymlinkMode         `protobuf:"varint,1,opt,name=symlinkMode,proto3,enum=sync.SymlinkMode" json:"symlinkMode,omitempty"`
	WatchMode              filesystem.WatchMode     `protobuf:"varint,21,opt,name=watchMode,proto3,enum=filesystem.WatchMode" json:"watchMode,omitempty"`
	WatchPollingInterval   uint32                   `protobuf:"varint,22,opt,name=watchPollingInterval,proto3" json:"watchPollingInterval,omitempty"`
	DefaultIgnores         []string                 `protobuf:"bytes,31,rep,name=defaultIgnores,proto3" json:"defaultIgnores,omitempty"`
	Ignores                []string                 `protobuf:"bytes,32,rep,name=ignores,proto3" json:"ignores,omitempty"`
	IgnoreVCSMode          sync.IgnoreVCSMode       `protobuf:"varint,33,opt,name=ignoreVCSMode,proto3,enum=sync.IgnoreVCSMode" json:"ignoreVCSMode,omitempty"`
	DefaultFileMode        uint32                   `protobuf:"varint,63,opt,name=defaultFileMode,proto3" json:"defaultFileMode,omitempty"`
	DefaultDirectoryMode   uint32                   `protobuf:"varint,64,opt,name=defaultDirectoryMode,proto3" json:"defaultDirectoryMode,omitempty"`
	DefaultOwner           string                   `protobuf:"bytes,65,opt,name=defaultOwner,proto3" json:"defaultOwner,omitempty"`
	DefaultGroup           string                   `protobuf:"bytes,66,opt,name=defaultGroup,proto3" json:"defaultGroup,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}                 `json:"-"`
	XXX_unrecognized       []byte                   `json:"-"`
	XXX_sizecache          int32                    `json:"-"`
}

func (m *Configuration) Reset()         { *m = Configuration{} }
func (m *Configuration) String() string { return proto.CompactTextString(m) }
func (*Configuration) ProtoMessage()    {}
func (*Configuration) Descriptor() ([]byte, []int) {
	return fileDescriptor_configuration_e35a36f0e41f48fe, []int{0}
}
func (m *Configuration) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Configuration.Unmarshal(m, b)
}
func (m *Configuration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Configuration.Marshal(b, m, deterministic)
}
func (dst *Configuration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Configuration.Merge(dst, src)
}
func (m *Configuration) XXX_Size() int {
	return xxx_messageInfo_Configuration.Size(m)
}
func (m *Configuration) XXX_DiscardUnknown() {
	xxx_messageInfo_Configuration.DiscardUnknown(m)
}

var xxx_messageInfo_Configuration proto.InternalMessageInfo

func (m *Configuration) GetSynchronizationMode() sync.SynchronizationMode {
	if m != nil {
		return m.SynchronizationMode
	}
	return sync.SynchronizationMode_SynchronizationModeDefault
}

func (m *Configuration) GetMaximumEntryCount() uint64 {
	if m != nil {
		return m.MaximumEntryCount
	}
	return 0
}

func (m *Configuration) GetMaximumStagingFileSize() uint64 {
	if m != nil {
		return m.MaximumStagingFileSize
	}
	return 0
}

func (m *Configuration) GetSymlinkMode() sync.SymlinkMode {
	if m != nil {
		return m.SymlinkMode
	}
	return sync.SymlinkMode_SymlinkDefault
}

func (m *Configuration) GetWatchMode() filesystem.WatchMode {
	if m != nil {
		return m.WatchMode
	}
	return filesystem.WatchMode_WatchModeDefault
}

func (m *Configuration) GetWatchPollingInterval() uint32 {
	if m != nil {
		return m.WatchPollingInterval
	}
	return 0
}

func (m *Configuration) GetDefaultIgnores() []string {
	if m != nil {
		return m.DefaultIgnores
	}
	return nil
}

func (m *Configuration) GetIgnores() []string {
	if m != nil {
		return m.Ignores
	}
	return nil
}

func (m *Configuration) GetIgnoreVCSMode() sync.IgnoreVCSMode {
	if m != nil {
		return m.IgnoreVCSMode
	}
	return sync.IgnoreVCSMode_IgnoreVCSDefault
}

func (m *Configuration) GetDefaultFileMode() uint32 {
	if m != nil {
		return m.DefaultFileMode
	}
	return 0
}

func (m *Configuration) GetDefaultDirectoryMode() uint32 {
	if m != nil {
		return m.DefaultDirectoryMode
	}
	return 0
}

func (m *Configuration) GetDefaultOwner() string {
	if m != nil {
		return m.DefaultOwner
	}
	return ""
}

func (m *Configuration) GetDefaultGroup() string {
	if m != nil {
		return m.DefaultGroup
	}
	return ""
}

func init() {
	proto.RegisterType((*Configuration)(nil), "session.Configuration")
}

func init() {
	proto.RegisterFile("session/configuration.proto", fileDescriptor_configuration_e35a36f0e41f48fe)
}

var fileDescriptor_configuration_e35a36f0e41f48fe = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0xdf, 0x6f, 0xd3, 0x30,
	0x10, 0xc7, 0x15, 0xf1, 0x63, 0xaa, 0xb7, 0x6e, 0xaa, 0xc7, 0xaa, 0x30, 0x1e, 0x08, 0x7b, 0x80,
	0x20, 0x41, 0x22, 0xad, 0x12, 0x12, 0x4f, 0xc0, 0xca, 0x0f, 0x55, 0x08, 0x81, 0x52, 0x09, 0x24,
	0xde, 0xbc, 0xd4, 0x75, 0x4f, 0x8b, 0x7d, 0x95, 0xed, 0x6c, 0x64, 0xaf, 0xfc, 0xe3, 0xa8, 0x17,
	0x97, 0xb5, 0x6b, 0xf7, 0x16, 0x7f, 0xbe, 0x9f, 0x93, 0xef, 0x2e, 0x09, 0x7b, 0xe2, 0xa4, 0x73,
	0x80, 0x26, 0x2f, 0xd1, 0x4c, 0x41, 0xd5, 0x56, 0x78, 0x40, 0x93, 0xcd, 0x2d, 0x7a, 0xe4, 0x3b,
	0x21, 0x3c, 0xee, 0x4f, 0xa1, 0x92, 0xae, 0x71, 0x5e, 0xea, 0xfc, 0x4a, 0xf8, 0x72, 0xd6, 0x0a,
	0xc7, 0x3d, 0xd7, 0x98, 0x32, 0x07, 0x65, 0xd0, 0xca, 0x80, 0x0e, 0x08, 0x69, 0x9c, 0x2c, 0x01,
	0x27, 0xe0, 0x1a, 0x5d, 0x81, 0xb9, 0x68, 0xd9, 0xc9, 0xdf, 0x07, 0xac, 0x3b, 0x5c, 0xbd, 0x90,
	0x7f, 0x65, 0x87, 0x0b, 0x6f, 0x66, 0xd1, 0xc0, 0x35, 0xa1, 0x6f, 0x38, 0x91, 0xf1, 0x6e, 0x12,
	0xa5, 0xfb, 0xa7, 0x8f, 0xb3, 0x45, 0x96, 0x8d, 0x37, 0x85, 0x62, 0x5b, 0x15, 0x7f, 0xc5, 0x7a,
	0x5a, 0xfc, 0x01, 0x5d, 0xeb, 0x4f, 0xc6, 0xdb, 0x66, 0x88, 0xb5, 0xf1, 0xf1, 0x5e, 0x12, 0xa5,
	0xf7, 0x8b, 0xcd, 0x80, 0xbf, 0x61, 0xfd, 0x00, 0xc7, 0x5e, 0x28, 0x30, 0xea, 0x33, 0x54, 0x72,
	0x0c, 0xd7, 0x32, 0xee, 0x52, 0xc9, 0x1d, 0x29, 0x1f, 0xb0, 0xdd, 0x30, 0x15, 0xb5, 0x1a, 0x51,
	0xab, 0xbd, 0x65, 0xab, 0xff, 0x83, 0x62, 0xd5, 0xe2, 0x03, 0xd6, 0xa1, 0x05, 0x52, 0xc9, 0x11,
	0x95, 0x1c, 0x65, 0x37, 0xdb, 0xcd, 0x7e, 0x2d, 0xc3, 0xe2, 0xc6, 0xe3, 0xa7, 0xec, 0x11, 0x1d,
	0x7e, 0x60, 0x55, 0x81, 0x51, 0x23, 0xe3, 0xa5, 0xbd, 0x14, 0x55, 0xdc, 0x4f, 0xa2, 0xb4, 0x5b,
	0x6c, 0xcd, 0xf8, 0x73, 0xb6, 0x3f, 0x91, 0x53, 0x51, 0x57, 0x7e, 0x44, 0xaf, 0xc7, 0xc5, 0x4f,
	0x93, 0x7b, 0x69, 0xa7, 0xb8, 0x45, 0x79, 0xcc, 0x76, 0x20, 0x08, 0x09, 0x09, 0xcb, 0x23, 0x7f,
	0xcb, 0xba, 0xed, 0xe3, 0xcf, 0xe1, 0x98, 0xda, 0x7d, 0x46, 0xed, 0x1e, 0xb6, 0x13, 0x8e, 0x56,
	0xa3, 0x62, 0xdd, 0xe4, 0x29, 0x3b, 0x08, 0xd7, 0x2c, 0xb6, 0x45, 0xc5, 0xef, 0xa8, 0xd7, 0xdb,
	0x78, 0x31, 0x5a, 0x40, 0x1f, 0xc1, 0xca, 0xd2, 0xa3, 0x6d, 0x48, 0x7f, 0xdf, 0x8e, 0xb6, 0x2d,
	0xe3, 0x27, 0x6c, 0x2f, 0xf0, 0xef, 0x57, 0x46, 0xda, 0xf8, 0x43, 0x12, 0xa5, 0x9d, 0x62, 0x8d,
	0xad, 0x38, 0x5f, 0x2c, 0xd6, 0xf3, 0xf8, 0x6c, 0xcd, 0x21, 0x76, 0xf6, 0xf2, 0xf7, 0x0b, 0x05,
	0x7e, 0x56, 0x9f, 0x67, 0x25, 0xea, 0x7c, 0x26, 0x2e, 0xb1, 0x7c, 0x0d, 0x98, 0xeb, 0xda, 0x0b,
	0x25, 0x4d, 0x3e, 0xbf, 0x50, 0x79, 0xf8, 0x01, 0xce, 0x1f, 0xd2, 0x77, 0x3b, 0xf8, 0x17, 0x00,
	0x00, 0xff, 0xff, 0xe8, 0xfa, 0x3f, 0x82, 0x2f, 0x03, 0x00, 0x00,
}
