// Code generated by protoc-gen-go. DO NOT EDIT.
// source: session/state.proto

package session // import "github.com/RokyErickson/doppelganger/pkg/session"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import rsync "github.com/RokyErickson/doppelganger/pkg/rsync"
import sync "github.com/RokyErickson/doppelganger/pkg/sync"

var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const _ = proto.ProtoPackageIsVersion2

type Status int32

const (
	Status_Disconnected           Status = 0
	Status_HaltedOnRootDeletion   Status = 1
	Status_HaltedOnRootTypeChange Status = 2
	Status_ConnectingAlpha        Status = 3
	Status_ConnectingBeta         Status = 4
	Status_Watching               Status = 5
	Status_Scanning               Status = 6
	Status_WaitingForRescan       Status = 7
	Status_Reconciling            Status = 8
	Status_StagingAlpha           Status = 9
	Status_StagingBeta            Status = 10
	Status_Transitioning          Status = 11
	Status_Saving                 Status = 12
)

var Status_name = map[int32]string{
	0:  "Disconnected",
	1:  "HaltedOnRootDeletion",
	2:  "HaltedOnRootTypeChange",
	3:  "ConnectingAlpha",
	4:  "ConnectingBeta",
	5:  "Watching",
	6:  "Scanning",
	7:  "WaitingForRescan",
	8:  "Reconciling",
	9:  "StagingAlpha",
	10: "StagingBeta",
	11: "Transitioning",
	12: "Saving",
}
var Status_value = map[string]int32{
	"Disconnected":           0,
	"HaltedOnRootDeletion":   1,
	"HaltedOnRootTypeChange": 2,
	"ConnectingAlpha":        3,
	"ConnectingBeta":         4,
	"Watching":               5,
	"Scanning":               6,
	"WaitingForRescan":       7,
	"Reconciling":            8,
	"StagingAlpha":           9,
	"StagingBeta":            10,
	"Transitioning":          11,
	"Saving":                 12,
}

func (x Status) String() string {
	return proto.EnumName(Status_name, int32(x))
}
func (Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_state_cddd0f3d09b59394, []int{0}
}

type State struct {
	Session                         *Session              `protobuf:"bytes,1,opt,name=session,proto3" json:"session,omitempty"`
	Status                          Status                `protobuf:"varint,2,opt,name=status,proto3,enum=session.Status" json:"status,omitempty"`
	AlphaConnected                  bool                  `protobuf:"varint,3,opt,name=alphaConnected,proto3" json:"alphaConnected,omitempty"`
	BetaConnected                   bool                  `protobuf:"varint,4,opt,name=betaConnected,proto3" json:"betaConnected,omitempty"`
	LastError                       string                `protobuf:"bytes,5,opt,name=lastError,proto3" json:"lastError,omitempty"`
	SuccessfulSynchronizationCycles uint64                `protobuf:"varint,6,opt,name=successfulSynchronizationCycles,proto3" json:"successfulSynchronizationCycles,omitempty"`
	StagingStatus                   *rsync.ReceiverStatus `protobuf:"bytes,7,opt,name=stagingStatus,proto3" json:"stagingStatus,omitempty"`
	Conflicts                       []*sync.Conflict      `protobuf:"bytes,8,rep,name=conflicts,proto3" json:"conflicts,omitempty"`
	AlphaProblems                   []*sync.Problem       `protobuf:"bytes,9,rep,name=alphaProblems,proto3" json:"alphaProblems,omitempty"`
	BetaProblems                    []*sync.Problem       `protobuf:"bytes,10,rep,name=betaProblems,proto3" json:"betaProblems,omitempty"`
	XXX_NoUnkeyedLiteral            struct{}              `json:"-"`
	XXX_unrecognized                []byte                `json:"-"`
	XXX_sizecache                   int32                 `json:"-"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_state_cddd0f3d09b59394, []int{0}
}
func (m *State) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_State.Unmarshal(m, b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_State.Marshal(b, m, deterministic)
}
func (dst *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(dst, src)
}
func (m *State) XXX_Size() int {
	return xxx_messageInfo_State.Size(m)
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetSession() *Session {
	if m != nil {
		return m.Session
	}
	return nil
}

func (m *State) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_Disconnected
}

func (m *State) GetAlphaConnected() bool {
	if m != nil {
		return m.AlphaConnected
	}
	return false
}

func (m *State) GetBetaConnected() bool {
	if m != nil {
		return m.BetaConnected
	}
	return false
}

func (m *State) GetLastError() string {
	if m != nil {
		return m.LastError
	}
	return ""
}

func (m *State) GetSuccessfulSynchronizationCycles() uint64 {
	if m != nil {
		return m.SuccessfulSynchronizationCycles
	}
	return 0
}

func (m *State) GetStagingStatus() *rsync.ReceiverStatus {
	if m != nil {
		return m.StagingStatus
	}
	return nil
}

func (m *State) GetConflicts() []*sync.Conflict {
	if m != nil {
		return m.Conflicts
	}
	return nil
}

func (m *State) GetAlphaProblems() []*sync.Problem {
	if m != nil {
		return m.AlphaProblems
	}
	return nil
}

func (m *State) GetBetaProblems() []*sync.Problem {
	if m != nil {
		return m.BetaProblems
	}
	return nil
}

func init() {
	proto.RegisterType((*State)(nil), "session.State")
	proto.RegisterEnum("session.Status", Status_name, Status_value)
}

func init() { proto.RegisterFile("session/state.proto", fileDescriptor_state_cddd0f3d09b59394) }

var fileDescriptor_state_cddd0f3d09b59394 = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0x5f, 0x6b, 0xdb, 0x3c,
	0x14, 0xc6, 0x5f, 0x37, 0x89, 0x93, 0x9c, 0xfc, 0xf3, 0x7b, 0xd2, 0x0e, 0x13, 0x06, 0x33, 0x63,
	0xac, 0x5e, 0xd9, 0x1c, 0x96, 0x5e, 0xee, 0x6a, 0x4d, 0x37, 0x7a, 0xb7, 0x61, 0x17, 0x0a, 0xbb,
	0x53, 0x54, 0xd5, 0x16, 0x73, 0xa4, 0x60, 0x29, 0x81, 0xec, 0xfb, 0xee, 0x6b, 0x8c, 0x21, 0x59,
	0x6e, 0x9a, 0x31, 0xd8, 0x95, 0xac, 0x47, 0xbf, 0xc7, 0xe7, 0xe8, 0x39, 0x08, 0xa6, 0x8a, 0x29,
	0xc5, 0xa5, 0x98, 0x2b, 0x4d, 0x34, 0x4b, 0x36, 0x95, 0xd4, 0x12, 0xbb, 0x4e, 0x9c, 0x4d, 0x2b,
	0xb5, 0x17, 0x74, 0x5e, 0x31, 0xca, 0xf8, 0xce, 0x9d, 0xce, 0xce, 0x1e, 0x2d, 0xf5, 0xea, 0xe4,
	0xa9, 0x45, 0xa9, 0x14, 0x0f, 0x25, 0xa7, 0xda, 0x89, 0x68, 0xc5, 0x4d, 0x25, 0x57, 0x25, 0x5b,
	0xd7, 0xda, 0xcb, 0x9f, 0x2d, 0xe8, 0x64, 0xa6, 0x1a, 0x5e, 0x40, 0x53, 0x29, 0xf4, 0x22, 0x2f,
	0x1e, 0x2c, 0x82, 0xa4, 0xf9, 0x67, 0x56, 0xaf, 0x69, 0x03, 0xe0, 0x39, 0xf8, 0xa6, 0xc5, 0xad,
	0x0a, 0x4f, 0x22, 0x2f, 0x1e, 0x2f, 0x26, 0x07, 0xd4, 0xca, 0xa9, 0x3b, 0xc6, 0xd7, 0x30, 0x26,
	0xe5, 0xa6, 0x20, 0x4b, 0x29, 0x04, 0xa3, 0x9a, 0xdd, 0x87, 0xad, 0xc8, 0x8b, 0x7b, 0xe9, 0x1f,
	0x2a, 0xbe, 0x82, 0xd1, 0x8a, 0xe9, 0x27, 0x58, 0xdb, 0x62, 0xc7, 0x22, 0x3e, 0x87, 0x7e, 0x49,
	0x94, 0xfe, 0x54, 0x55, 0xb2, 0x0a, 0x3b, 0x91, 0x17, 0xf7, 0xd3, 0x83, 0x80, 0x37, 0xf0, 0x42,
	0x6d, 0x29, 0x65, 0x4a, 0x3d, 0x6c, 0xcb, 0x6c, 0x2f, 0x68, 0x51, 0x49, 0xc1, 0x7f, 0x10, 0xcd,
	0xa5, 0x58, 0xee, 0x69, 0xc9, 0x54, 0xe8, 0x47, 0x5e, 0xdc, 0x4e, 0xff, 0x85, 0xe1, 0x07, 0x18,
	0x29, 0x4d, 0x72, 0x2e, 0xf2, 0xfa, 0x3a, 0x61, 0xd7, 0x06, 0x72, 0x96, 0xd8, 0x09, 0x24, 0x69,
	0x3d, 0x81, 0xca, 0xdd, 0xf5, 0x98, 0xc5, 0xb7, 0xd0, 0x6f, 0x72, 0x57, 0x61, 0x2f, 0x6a, 0xc5,
	0x83, 0xc5, 0x38, 0xb1, 0xbe, 0xa5, 0x93, 0xd3, 0x03, 0x80, 0x97, 0x30, 0xb2, 0x51, 0x7c, 0xad,
	0xa7, 0xa2, 0xc2, 0xbe, 0x75, 0x8c, 0x6a, 0x87, 0x53, 0xd3, 0x63, 0x06, 0xdf, 0xc3, 0xd0, 0x04,
	0xf3, 0xe8, 0x81, 0xbf, 0x79, 0x8e, 0x90, 0x8b, 0x5f, 0x1e, 0xf8, 0xae, 0xc1, 0x00, 0x86, 0xd7,
	0x5c, 0xd1, 0x26, 0xd5, 0xe0, 0x3f, 0x0c, 0xe1, 0xf4, 0x86, 0x94, 0x9a, 0xdd, 0x7f, 0x11, 0xa9,
	0x94, 0xfa, 0x9a, 0x95, 0xcc, 0xa4, 0x11, 0x78, 0x38, 0x83, 0x67, 0x4f, 0x4f, 0x6e, 0xf7, 0x1b,
	0xb6, 0x2c, 0x88, 0xc8, 0x59, 0x70, 0x82, 0x53, 0x98, 0xb8, 0xd1, 0x70, 0x91, 0x7f, 0x34, 0x0d,
	0x06, 0x2d, 0x44, 0x18, 0x1f, 0xc4, 0x2b, 0xa6, 0x49, 0xd0, 0xc6, 0x21, 0xf4, 0xee, 0x88, 0xa6,
	0x05, 0x17, 0x79, 0xd0, 0x31, 0xbb, 0x8c, 0x12, 0x21, 0xcc, 0xce, 0xc7, 0x53, 0x08, 0xee, 0x08,
	0x37, 0xf0, 0x67, 0x59, 0xa5, 0x4c, 0x51, 0x22, 0x82, 0x2e, 0x4e, 0x60, 0x90, 0x32, 0x2a, 0x05,
	0xe5, 0xa5, 0xc1, 0x7a, 0xa6, 0xe7, 0xac, 0x4e, 0xb9, 0x2e, 0xd4, 0x37, 0x88, 0x53, 0x6c, 0x15,
	0xc0, 0xff, 0x61, 0x74, 0x5b, 0x11, 0xa1, 0xb8, 0x69, 0xdd, 0xb8, 0x06, 0x08, 0xe0, 0x67, 0x64,
	0x67, 0xbe, 0x87, 0x57, 0x6f, 0xbe, 0x9d, 0xe7, 0x5c, 0x17, 0xdb, 0x55, 0x42, 0xe5, 0x7a, 0x5e,
	0x90, 0x9d, 0xa4, 0xef, 0xb8, 0x9c, 0xaf, 0xb7, 0x9a, 0xe4, 0x4c, 0xcc, 0x37, 0xdf, 0xf3, 0xe6,
	0x09, 0xad, 0x7c, 0xfb, 0x34, 0x2e, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xb7, 0x87, 0x0b, 0xb9,
	0x8f, 0x03, 0x00, 0x00,
}