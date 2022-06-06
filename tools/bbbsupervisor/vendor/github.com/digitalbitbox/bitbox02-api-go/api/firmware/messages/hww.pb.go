// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hww.proto

package messages

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Error struct {
	Code                 int32    `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_c12bd9e6412b8a05, []int{0}
}

func (m *Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Error.Unmarshal(m, b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Error.Marshal(b, m, deterministic)
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return xxx_messageInfo_Error.Size(m)
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type Success struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Success) Reset()         { *m = Success{} }
func (m *Success) String() string { return proto.CompactTextString(m) }
func (*Success) ProtoMessage()    {}
func (*Success) Descriptor() ([]byte, []int) {
	return fileDescriptor_c12bd9e6412b8a05, []int{1}
}

func (m *Success) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Success.Unmarshal(m, b)
}
func (m *Success) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Success.Marshal(b, m, deterministic)
}
func (m *Success) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Success.Merge(m, src)
}
func (m *Success) XXX_Size() int {
	return xxx_messageInfo_Success.Size(m)
}
func (m *Success) XXX_DiscardUnknown() {
	xxx_messageInfo_Success.DiscardUnknown(m)
}

var xxx_messageInfo_Success proto.InternalMessageInfo

type Request struct {
	// Types that are valid to be assigned to Request:
	//	*Request_RandomNumber
	//	*Request_DeviceName
	//	*Request_DeviceLanguage
	//	*Request_DeviceInfo
	//	*Request_SetPassword
	//	*Request_CreateBackup
	//	*Request_ShowMnemonic
	//	*Request_BtcPub
	//	*Request_BtcSignInit
	//	*Request_BtcSignInput
	//	*Request_BtcSignOutput
	//	*Request_InsertRemoveSdcard
	//	*Request_CheckSdcard
	//	*Request_SetMnemonicPassphraseEnabled
	//	*Request_ListBackups
	//	*Request_RestoreBackup
	//	*Request_PerformAttestation
	//	*Request_Reboot
	//	*Request_CheckBackup
	//	*Request_Eth
	//	*Request_Reset_
	//	*Request_RestoreFromMnemonic
	//	*Request_Bitboxbase
	Request              isRequest_Request `protobuf_oneof:"request"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_c12bd9e6412b8a05, []int{2}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

type isRequest_Request interface {
	isRequest_Request()
}

type Request_RandomNumber struct {
	RandomNumber *RandomNumberRequest `protobuf:"bytes,1,opt,name=random_number,json=randomNumber,proto3,oneof"`
}

type Request_DeviceName struct {
	DeviceName *SetDeviceNameRequest `protobuf:"bytes,2,opt,name=device_name,json=deviceName,proto3,oneof"`
}

type Request_DeviceLanguage struct {
	DeviceLanguage *SetDeviceLanguageRequest `protobuf:"bytes,3,opt,name=device_language,json=deviceLanguage,proto3,oneof"`
}

type Request_DeviceInfo struct {
	DeviceInfo *DeviceInfoRequest `protobuf:"bytes,4,opt,name=device_info,json=deviceInfo,proto3,oneof"`
}

type Request_SetPassword struct {
	SetPassword *SetPasswordRequest `protobuf:"bytes,5,opt,name=set_password,json=setPassword,proto3,oneof"`
}

type Request_CreateBackup struct {
	CreateBackup *CreateBackupRequest `protobuf:"bytes,6,opt,name=create_backup,json=createBackup,proto3,oneof"`
}

type Request_ShowMnemonic struct {
	ShowMnemonic *ShowMnemonicRequest `protobuf:"bytes,7,opt,name=show_mnemonic,json=showMnemonic,proto3,oneof"`
}

type Request_BtcPub struct {
	BtcPub *BTCPubRequest `protobuf:"bytes,8,opt,name=btc_pub,json=btcPub,proto3,oneof"`
}

type Request_BtcSignInit struct {
	BtcSignInit *BTCSignInitRequest `protobuf:"bytes,9,opt,name=btc_sign_init,json=btcSignInit,proto3,oneof"`
}

type Request_BtcSignInput struct {
	BtcSignInput *BTCSignInputRequest `protobuf:"bytes,10,opt,name=btc_sign_input,json=btcSignInput,proto3,oneof"`
}

type Request_BtcSignOutput struct {
	BtcSignOutput *BTCSignOutputRequest `protobuf:"bytes,11,opt,name=btc_sign_output,json=btcSignOutput,proto3,oneof"`
}

type Request_InsertRemoveSdcard struct {
	InsertRemoveSdcard *InsertRemoveSDCardRequest `protobuf:"bytes,12,opt,name=insert_remove_sdcard,json=insertRemoveSdcard,proto3,oneof"`
}

type Request_CheckSdcard struct {
	CheckSdcard *CheckSDCardRequest `protobuf:"bytes,13,opt,name=check_sdcard,json=checkSdcard,proto3,oneof"`
}

type Request_SetMnemonicPassphraseEnabled struct {
	SetMnemonicPassphraseEnabled *SetMnemonicPassphraseEnabledRequest `protobuf:"bytes,14,opt,name=set_mnemonic_passphrase_enabled,json=setMnemonicPassphraseEnabled,proto3,oneof"`
}

type Request_ListBackups struct {
	ListBackups *ListBackupsRequest `protobuf:"bytes,15,opt,name=list_backups,json=listBackups,proto3,oneof"`
}

type Request_RestoreBackup struct {
	RestoreBackup *RestoreBackupRequest `protobuf:"bytes,16,opt,name=restore_backup,json=restoreBackup,proto3,oneof"`
}

type Request_PerformAttestation struct {
	PerformAttestation *PerformAttestationRequest `protobuf:"bytes,17,opt,name=perform_attestation,json=performAttestation,proto3,oneof"`
}

type Request_Reboot struct {
	Reboot *RebootRequest `protobuf:"bytes,18,opt,name=reboot,proto3,oneof"`
}

type Request_CheckBackup struct {
	CheckBackup *CheckBackupRequest `protobuf:"bytes,19,opt,name=check_backup,json=checkBackup,proto3,oneof"`
}

type Request_Eth struct {
	Eth *ETHRequest `protobuf:"bytes,20,opt,name=eth,proto3,oneof"`
}

type Request_Reset_ struct {
	Reset_ *ResetRequest `protobuf:"bytes,21,opt,name=reset,proto3,oneof"`
}

type Request_RestoreFromMnemonic struct {
	RestoreFromMnemonic *RestoreFromMnemonicRequest `protobuf:"bytes,22,opt,name=restore_from_mnemonic,json=restoreFromMnemonic,proto3,oneof"`
}

type Request_Bitboxbase struct {
	Bitboxbase *BitBoxBaseRequest `protobuf:"bytes,23,opt,name=bitboxbase,proto3,oneof"`
}

func (*Request_RandomNumber) isRequest_Request() {}

func (*Request_DeviceName) isRequest_Request() {}

func (*Request_DeviceLanguage) isRequest_Request() {}

func (*Request_DeviceInfo) isRequest_Request() {}

func (*Request_SetPassword) isRequest_Request() {}

func (*Request_CreateBackup) isRequest_Request() {}

func (*Request_ShowMnemonic) isRequest_Request() {}

func (*Request_BtcPub) isRequest_Request() {}

func (*Request_BtcSignInit) isRequest_Request() {}

func (*Request_BtcSignInput) isRequest_Request() {}

func (*Request_BtcSignOutput) isRequest_Request() {}

func (*Request_InsertRemoveSdcard) isRequest_Request() {}

func (*Request_CheckSdcard) isRequest_Request() {}

func (*Request_SetMnemonicPassphraseEnabled) isRequest_Request() {}

func (*Request_ListBackups) isRequest_Request() {}

func (*Request_RestoreBackup) isRequest_Request() {}

func (*Request_PerformAttestation) isRequest_Request() {}

func (*Request_Reboot) isRequest_Request() {}

func (*Request_CheckBackup) isRequest_Request() {}

func (*Request_Eth) isRequest_Request() {}

func (*Request_Reset_) isRequest_Request() {}

func (*Request_RestoreFromMnemonic) isRequest_Request() {}

func (*Request_Bitboxbase) isRequest_Request() {}

func (m *Request) GetRequest() isRequest_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *Request) GetRandomNumber() *RandomNumberRequest {
	if x, ok := m.GetRequest().(*Request_RandomNumber); ok {
		return x.RandomNumber
	}
	return nil
}

func (m *Request) GetDeviceName() *SetDeviceNameRequest {
	if x, ok := m.GetRequest().(*Request_DeviceName); ok {
		return x.DeviceName
	}
	return nil
}

func (m *Request) GetDeviceLanguage() *SetDeviceLanguageRequest {
	if x, ok := m.GetRequest().(*Request_DeviceLanguage); ok {
		return x.DeviceLanguage
	}
	return nil
}

func (m *Request) GetDeviceInfo() *DeviceInfoRequest {
	if x, ok := m.GetRequest().(*Request_DeviceInfo); ok {
		return x.DeviceInfo
	}
	return nil
}

func (m *Request) GetSetPassword() *SetPasswordRequest {
	if x, ok := m.GetRequest().(*Request_SetPassword); ok {
		return x.SetPassword
	}
	return nil
}

func (m *Request) GetCreateBackup() *CreateBackupRequest {
	if x, ok := m.GetRequest().(*Request_CreateBackup); ok {
		return x.CreateBackup
	}
	return nil
}

func (m *Request) GetShowMnemonic() *ShowMnemonicRequest {
	if x, ok := m.GetRequest().(*Request_ShowMnemonic); ok {
		return x.ShowMnemonic
	}
	return nil
}

func (m *Request) GetBtcPub() *BTCPubRequest {
	if x, ok := m.GetRequest().(*Request_BtcPub); ok {
		return x.BtcPub
	}
	return nil
}

func (m *Request) GetBtcSignInit() *BTCSignInitRequest {
	if x, ok := m.GetRequest().(*Request_BtcSignInit); ok {
		return x.BtcSignInit
	}
	return nil
}

func (m *Request) GetBtcSignInput() *BTCSignInputRequest {
	if x, ok := m.GetRequest().(*Request_BtcSignInput); ok {
		return x.BtcSignInput
	}
	return nil
}

func (m *Request) GetBtcSignOutput() *BTCSignOutputRequest {
	if x, ok := m.GetRequest().(*Request_BtcSignOutput); ok {
		return x.BtcSignOutput
	}
	return nil
}

func (m *Request) GetInsertRemoveSdcard() *InsertRemoveSDCardRequest {
	if x, ok := m.GetRequest().(*Request_InsertRemoveSdcard); ok {
		return x.InsertRemoveSdcard
	}
	return nil
}

func (m *Request) GetCheckSdcard() *CheckSDCardRequest {
	if x, ok := m.GetRequest().(*Request_CheckSdcard); ok {
		return x.CheckSdcard
	}
	return nil
}

func (m *Request) GetSetMnemonicPassphraseEnabled() *SetMnemonicPassphraseEnabledRequest {
	if x, ok := m.GetRequest().(*Request_SetMnemonicPassphraseEnabled); ok {
		return x.SetMnemonicPassphraseEnabled
	}
	return nil
}

func (m *Request) GetListBackups() *ListBackupsRequest {
	if x, ok := m.GetRequest().(*Request_ListBackups); ok {
		return x.ListBackups
	}
	return nil
}

func (m *Request) GetRestoreBackup() *RestoreBackupRequest {
	if x, ok := m.GetRequest().(*Request_RestoreBackup); ok {
		return x.RestoreBackup
	}
	return nil
}

func (m *Request) GetPerformAttestation() *PerformAttestationRequest {
	if x, ok := m.GetRequest().(*Request_PerformAttestation); ok {
		return x.PerformAttestation
	}
	return nil
}

func (m *Request) GetReboot() *RebootRequest {
	if x, ok := m.GetRequest().(*Request_Reboot); ok {
		return x.Reboot
	}
	return nil
}

func (m *Request) GetCheckBackup() *CheckBackupRequest {
	if x, ok := m.GetRequest().(*Request_CheckBackup); ok {
		return x.CheckBackup
	}
	return nil
}

func (m *Request) GetEth() *ETHRequest {
	if x, ok := m.GetRequest().(*Request_Eth); ok {
		return x.Eth
	}
	return nil
}

func (m *Request) GetReset_() *ResetRequest {
	if x, ok := m.GetRequest().(*Request_Reset_); ok {
		return x.Reset_
	}
	return nil
}

func (m *Request) GetRestoreFromMnemonic() *RestoreFromMnemonicRequest {
	if x, ok := m.GetRequest().(*Request_RestoreFromMnemonic); ok {
		return x.RestoreFromMnemonic
	}
	return nil
}

func (m *Request) GetBitboxbase() *BitBoxBaseRequest {
	if x, ok := m.GetRequest().(*Request_Bitboxbase); ok {
		return x.Bitboxbase
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Request) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Request_RandomNumber)(nil),
		(*Request_DeviceName)(nil),
		(*Request_DeviceLanguage)(nil),
		(*Request_DeviceInfo)(nil),
		(*Request_SetPassword)(nil),
		(*Request_CreateBackup)(nil),
		(*Request_ShowMnemonic)(nil),
		(*Request_BtcPub)(nil),
		(*Request_BtcSignInit)(nil),
		(*Request_BtcSignInput)(nil),
		(*Request_BtcSignOutput)(nil),
		(*Request_InsertRemoveSdcard)(nil),
		(*Request_CheckSdcard)(nil),
		(*Request_SetMnemonicPassphraseEnabled)(nil),
		(*Request_ListBackups)(nil),
		(*Request_RestoreBackup)(nil),
		(*Request_PerformAttestation)(nil),
		(*Request_Reboot)(nil),
		(*Request_CheckBackup)(nil),
		(*Request_Eth)(nil),
		(*Request_Reset_)(nil),
		(*Request_RestoreFromMnemonic)(nil),
		(*Request_Bitboxbase)(nil),
	}
}

type Response struct {
	// Types that are valid to be assigned to Response:
	//	*Response_Success
	//	*Response_Error
	//	*Response_RandomNumber
	//	*Response_DeviceInfo
	//	*Response_Pub
	//	*Response_BtcSignNext
	//	*Response_ListBackups
	//	*Response_CheckBackup
	//	*Response_PerformAttestation
	//	*Response_CheckSdcard
	//	*Response_Eth
	Response             isResponse_Response `protobuf_oneof:"response"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_c12bd9e6412b8a05, []int{3}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

type isResponse_Response interface {
	isResponse_Response()
}

type Response_Success struct {
	Success *Success `protobuf:"bytes,1,opt,name=success,proto3,oneof"`
}

type Response_Error struct {
	Error *Error `protobuf:"bytes,2,opt,name=error,proto3,oneof"`
}

type Response_RandomNumber struct {
	RandomNumber *RandomNumberResponse `protobuf:"bytes,3,opt,name=random_number,json=randomNumber,proto3,oneof"`
}

type Response_DeviceInfo struct {
	DeviceInfo *DeviceInfoResponse `protobuf:"bytes,4,opt,name=device_info,json=deviceInfo,proto3,oneof"`
}

type Response_Pub struct {
	Pub *PubResponse `protobuf:"bytes,5,opt,name=pub,proto3,oneof"`
}

type Response_BtcSignNext struct {
	BtcSignNext *BTCSignNextResponse `protobuf:"bytes,6,opt,name=btc_sign_next,json=btcSignNext,proto3,oneof"`
}

type Response_ListBackups struct {
	ListBackups *ListBackupsResponse `protobuf:"bytes,7,opt,name=list_backups,json=listBackups,proto3,oneof"`
}

type Response_CheckBackup struct {
	CheckBackup *CheckBackupResponse `protobuf:"bytes,8,opt,name=check_backup,json=checkBackup,proto3,oneof"`
}

type Response_PerformAttestation struct {
	PerformAttestation *PerformAttestationResponse `protobuf:"bytes,9,opt,name=perform_attestation,json=performAttestation,proto3,oneof"`
}

type Response_CheckSdcard struct {
	CheckSdcard *CheckSDCardResponse `protobuf:"bytes,10,opt,name=check_sdcard,json=checkSdcard,proto3,oneof"`
}

type Response_Eth struct {
	Eth *ETHResponse `protobuf:"bytes,11,opt,name=eth,proto3,oneof"`
}

func (*Response_Success) isResponse_Response() {}

func (*Response_Error) isResponse_Response() {}

func (*Response_RandomNumber) isResponse_Response() {}

func (*Response_DeviceInfo) isResponse_Response() {}

func (*Response_Pub) isResponse_Response() {}

func (*Response_BtcSignNext) isResponse_Response() {}

func (*Response_ListBackups) isResponse_Response() {}

func (*Response_CheckBackup) isResponse_Response() {}

func (*Response_PerformAttestation) isResponse_Response() {}

func (*Response_CheckSdcard) isResponse_Response() {}

func (*Response_Eth) isResponse_Response() {}

func (m *Response) GetResponse() isResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *Response) GetSuccess() *Success {
	if x, ok := m.GetResponse().(*Response_Success); ok {
		return x.Success
	}
	return nil
}

func (m *Response) GetError() *Error {
	if x, ok := m.GetResponse().(*Response_Error); ok {
		return x.Error
	}
	return nil
}

func (m *Response) GetRandomNumber() *RandomNumberResponse {
	if x, ok := m.GetResponse().(*Response_RandomNumber); ok {
		return x.RandomNumber
	}
	return nil
}

func (m *Response) GetDeviceInfo() *DeviceInfoResponse {
	if x, ok := m.GetResponse().(*Response_DeviceInfo); ok {
		return x.DeviceInfo
	}
	return nil
}

func (m *Response) GetPub() *PubResponse {
	if x, ok := m.GetResponse().(*Response_Pub); ok {
		return x.Pub
	}
	return nil
}

func (m *Response) GetBtcSignNext() *BTCSignNextResponse {
	if x, ok := m.GetResponse().(*Response_BtcSignNext); ok {
		return x.BtcSignNext
	}
	return nil
}

func (m *Response) GetListBackups() *ListBackupsResponse {
	if x, ok := m.GetResponse().(*Response_ListBackups); ok {
		return x.ListBackups
	}
	return nil
}

func (m *Response) GetCheckBackup() *CheckBackupResponse {
	if x, ok := m.GetResponse().(*Response_CheckBackup); ok {
		return x.CheckBackup
	}
	return nil
}

func (m *Response) GetPerformAttestation() *PerformAttestationResponse {
	if x, ok := m.GetResponse().(*Response_PerformAttestation); ok {
		return x.PerformAttestation
	}
	return nil
}

func (m *Response) GetCheckSdcard() *CheckSDCardResponse {
	if x, ok := m.GetResponse().(*Response_CheckSdcard); ok {
		return x.CheckSdcard
	}
	return nil
}

func (m *Response) GetEth() *ETHResponse {
	if x, ok := m.GetResponse().(*Response_Eth); ok {
		return x.Eth
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Response) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Response_Success)(nil),
		(*Response_Error)(nil),
		(*Response_RandomNumber)(nil),
		(*Response_DeviceInfo)(nil),
		(*Response_Pub)(nil),
		(*Response_BtcSignNext)(nil),
		(*Response_ListBackups)(nil),
		(*Response_CheckBackup)(nil),
		(*Response_PerformAttestation)(nil),
		(*Response_CheckSdcard)(nil),
		(*Response_Eth)(nil),
	}
}

func init() {
	proto.RegisterType((*Error)(nil), "Error")
	proto.RegisterType((*Success)(nil), "Success")
	proto.RegisterType((*Request)(nil), "Request")
	proto.RegisterType((*Response)(nil), "Response")
}

func init() { proto.RegisterFile("hww.proto", fileDescriptor_c12bd9e6412b8a05) }

var fileDescriptor_c12bd9e6412b8a05 = []byte{
	// 937 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x56, 0x6b, 0x6f, 0x1c, 0x35,
	0x14, 0x6d, 0x48, 0xf6, 0xe5, 0x7d, 0xa4, 0x78, 0x13, 0x70, 0x0b, 0xa2, 0x51, 0x54, 0xa4, 0xf0,
	0x65, 0x84, 0x02, 0x45, 0x0d, 0x54, 0x20, 0x76, 0x13, 0x94, 0x48, 0xed, 0xb2, 0xcc, 0xe6, 0xfb,
	0x68, 0x1e, 0x37, 0xd9, 0x51, 0x33, 0xf6, 0x60, 0x7b, 0xb2, 0xe1, 0x9f, 0xf0, 0xbf, 0xf8, 0x43,
	0xc8, 0x1e, 0x7b, 0xec, 0x7d, 0xa8, 0xdf, 0x76, 0xce, 0xbd, 0xe7, 0xfa, 0xce, 0xf5, 0xb9, 0x67,
	0x16, 0xf5, 0x96, 0xab, 0x55, 0x50, 0x72, 0x26, 0xd9, 0xcb, 0x41, 0xca, 0x8a, 0x82, 0x51, 0xf3,
	0x74, 0x9c, 0xc4, 0xe9, 0xc7, 0xaa, 0x8c, 0x14, 0x18, 0xd3, 0x4c, 0x34, 0x70, 0x2e, 0x13, 0xf6,
	0xf4, 0xfd, 0x79, 0x24, 0xfe, 0x11, 0x12, 0x0a, 0x03, 0x3f, 0xaf, 0xe1, 0x24, 0x16, 0x60, 0x90,
	0x5e, 0x22, 0x53, 0xfb, 0x13, 0xe4, 0xd2, 0xfc, 0x1c, 0x15, 0x14, 0x0a, 0x46, 0x73, 0x1b, 0x1a,
	0xf3, 0x98, 0x66, 0xac, 0x88, 0x68, 0x55, 0x24, 0xc0, 0x6d, 0x23, 0x6b, 0xa5, 0x5f, 0x94, 0xc0,
	0xef, 0x18, 0x2f, 0xa2, 0x58, 0x4a, 0x10, 0x32, 0x96, 0xb9, 0xed, 0xf1, 0xf4, 0x0d, 0x6a, 0x5d,
	0x71, 0xce, 0x38, 0xc6, 0xe8, 0x20, 0x65, 0x19, 0x90, 0xbd, 0x93, 0xbd, 0xb3, 0x56, 0xa8, 0x7f,
	0x63, 0x82, 0x3a, 0x05, 0x08, 0x11, 0xdf, 0x03, 0xf9, 0xec, 0x64, 0xef, 0xac, 0x17, 0xda, 0xc7,
	0xd3, 0x1e, 0xea, 0x2c, 0xaa, 0x34, 0x05, 0x21, 0x4e, 0xff, 0xed, 0xa3, 0x4e, 0x08, 0x7f, 0x57,
	0x20, 0x24, 0xfe, 0x05, 0x0d, 0xd7, 0xba, 0xd1, 0xd5, 0xfa, 0xe7, 0x47, 0x41, 0xa8, 0xd1, 0x99,
	0x06, 0x4d, 0xf2, 0xf5, 0xb3, 0x70, 0xc0, 0x3d, 0x18, 0xbf, 0x45, 0xfd, 0x0c, 0x1e, 0xf3, 0x14,
	0x22, 0x1a, 0x17, 0xf5, 0x89, 0xfd, 0xf3, 0xe3, 0x60, 0x01, 0xf2, 0x52, 0xc3, 0xb3, 0xb8, 0x00,
	0xc7, 0x45, 0x59, 0x03, 0xe2, 0x4b, 0x74, 0x68, 0x98, 0x0f, 0x31, 0xbd, 0xaf, 0x54, 0xbf, 0xfb,
	0x9a, 0xfd, 0xc2, 0xb1, 0xdf, 0x9b, 0x88, 0xab, 0x30, 0xca, 0xd6, 0x02, 0xf8, 0x4d, 0x73, 0x7e,
	0x4e, 0xef, 0x18, 0x39, 0xd0, 0x15, 0x70, 0x50, 0xd3, 0x6f, 0xe8, 0x1d, 0xdb, 0x3a, 0x5c, 0x81,
	0xf8, 0x2d, 0x1a, 0x08, 0x90, 0x51, 0x19, 0x0b, 0xb1, 0x62, 0x3c, 0x23, 0x2d, 0xcd, 0x1b, 0xab,
	0x93, 0xe7, 0x06, 0x73, 0xc4, 0xbe, 0x70, 0xa8, 0x9a, 0x56, 0xca, 0x21, 0x96, 0x10, 0xd5, 0x42,
	0x21, 0x6d, 0x33, 0xad, 0xa9, 0x46, 0x27, 0x1a, 0xf4, 0xa6, 0x95, 0x7a, 0xb0, 0x22, 0x8b, 0x25,
	0x5b, 0x45, 0x56, 0x0d, 0xa4, 0x63, 0xc8, 0x8b, 0x25, 0x5b, 0x7d, 0x30, 0xa0, 0x47, 0x16, 0x1e,
	0x8c, 0xbf, 0x43, 0x9d, 0x44, 0xa6, 0x51, 0x59, 0x25, 0xa4, 0xab, 0x69, 0xa3, 0x60, 0x72, 0x3b,
	0x9d, 0x57, 0x89, 0x23, 0xb4, 0x13, 0x99, 0xce, 0xab, 0x04, 0x5f, 0xa0, 0xa1, 0x4a, 0x15, 0xf9,
	0x3d, 0x8d, 0x72, 0x9a, 0x4b, 0xd2, 0x33, 0xef, 0x37, 0xb9, 0x9d, 0x2e, 0xf2, 0x7b, 0x7a, 0x43,
	0x73, 0xe9, 0xbd, 0x5f, 0x22, 0x53, 0x8b, 0xe2, 0x77, 0x68, 0xe4, 0x51, 0xcb, 0x4a, 0x12, 0x64,
	0x7a, 0x6c, 0xb8, 0x65, 0xe5, 0x91, 0x07, 0x0d, 0xb9, 0xac, 0x24, 0xfe, 0x0d, 0x1d, 0x36, 0x6c,
	0x56, 0x49, 0x45, 0xef, 0x1b, 0x49, 0x18, 0xfa, 0x9f, 0x1a, 0x75, 0xfc, 0xa1, 0xe1, 0xd7, 0x38,
	0x9e, 0xa1, 0xa3, 0x9c, 0x0a, 0xe0, 0x32, 0xe2, 0x50, 0xb0, 0x47, 0x88, 0x44, 0x96, 0xc6, 0x3c,
	0x23, 0x03, 0x5d, 0xe5, 0x65, 0x70, 0xa3, 0x83, 0xa1, 0x8e, 0x2d, 0x2e, 0xa7, 0xb1, 0x7f, 0x4f,
	0x38, 0xf7, 0x83, 0x9a, 0xa7, 0x2e, 0x3a, 0x5d, 0x42, 0xfa, 0xd1, 0xd6, 0x19, 0x9a, 0x41, 0x4c,
	0x15, 0xb8, 0x59, 0xa0, 0xaf, 0x53, 0x0d, 0xb3, 0x40, 0xaf, 0x94, 0x44, 0xec, 0x55, 0x69, 0xad,
	0x94, 0x4b, 0x1e, 0x0b, 0x88, 0x80, 0xc6, 0xc9, 0x03, 0x64, 0x64, 0xa4, 0x8b, 0xbd, 0x56, 0xaa,
	0xb1, 0xb7, 0x34, 0x6f, 0xb2, 0xae, 0xea, 0x24, 0x57, 0xfd, 0x6b, 0xf1, 0x89, 0x34, 0xd5, 0xe8,
	0x43, 0x2e, 0xa4, 0x51, 0x95, 0x20, 0x87, 0xa6, 0xd1, 0xf7, 0xb9, 0x90, 0xb5, 0x7a, 0x84, 0xd7,
	0xe8, 0x83, 0x43, 0xf1, 0xaf, 0x68, 0xc4, 0x41, 0x48, 0xc6, 0x1b, 0x49, 0x3e, 0x37, 0x23, 0x0f,
	0x6b, 0x78, 0x53, 0x93, 0x43, 0xee, 0xe3, 0xf8, 0x03, 0x1a, 0xef, 0xb0, 0x1a, 0xf2, 0xb9, 0x99,
	0xf8, 0xbc, 0x8e, 0xfd, 0xee, 0x42, 0xde, 0xc4, 0xcb, 0xad, 0x20, 0x3e, 0x43, 0x6d, 0x0e, 0x09,
	0x63, 0x92, 0x60, 0xa3, 0xd2, 0x50, 0x3f, 0x7a, 0x2a, 0xad, 0xe3, 0xee, 0x6e, 0x4c, 0xdb, 0x63,
	0xff, 0x6e, 0x36, 0x9b, 0xae, 0xef, 0xc6, 0xb4, 0xfc, 0x0a, 0xed, 0x83, 0x5c, 0x92, 0x23, 0x4d,
	0xe8, 0x07, 0x57, 0xb7, 0xd7, 0x2e, 0x51, 0x45, 0xf0, 0xb7, 0xa8, 0xc5, 0x41, 0x80, 0x24, 0xc7,
	0x3a, 0x65, 0xa8, 0x46, 0x01, 0x5e, 0x0b, 0x75, 0x14, 0xff, 0x85, 0x8e, 0xed, 0xe8, 0xee, 0x38,
	0x2b, 0xdc, 0x5e, 0x7e, 0xa1, 0x69, 0x5f, 0xd9, 0x09, 0xfe, 0xc1, 0x59, 0xb1, 0xbd, 0x9e, 0x63,
	0xbe, 0x1d, 0xc5, 0x3f, 0x22, 0xe4, 0xbe, 0x09, 0xe4, 0x4b, 0xe3, 0x47, 0x93, 0x5c, 0x4e, 0xd8,
	0xd3, 0x24, 0x16, 0xbe, 0x19, 0xba, 0xbc, 0x49, 0x0f, 0x75, 0x78, 0x1d, 0x38, 0xfd, 0xef, 0x00,
	0x75, 0x43, 0x10, 0x25, 0xa3, 0x02, 0xf0, 0x6b, 0xd4, 0x11, 0xb5, 0x65, 0x1b, 0x57, 0xee, 0x06,
	0xc6, 0xc2, 0xaf, 0x9f, 0x85, 0x36, 0x84, 0xbf, 0x41, 0x2d, 0x50, 0xdf, 0x03, 0x63, 0xbf, 0xed,
	0x40, 0x7f, 0x1d, 0xd4, 0x6b, 0x6a, 0x18, 0xbf, 0xdb, 0x74, 0xf8, 0x7d, 0x2b, 0x90, 0x35, 0x87,
	0xaf, 0xcf, 0xdc, 0xb2, 0xf8, 0x9f, 0x76, 0x59, 0xec, 0x78, 0xcd, 0x62, 0x1b, 0xa6, 0xef, 0xb1,
	0x27, 0x68, 0x5f, 0x79, 0x55, 0x6d, 0xad, 0x83, 0x40, 0x1b, 0x55, 0x93, 0xa8, 0x42, 0xf8, 0x67,
	0xcf, 0xa6, 0x28, 0x3c, 0xc9, 0xc6, 0x4b, 0x8d, 0x57, 0xcc, 0xe0, 0x49, 0x7a, 0x1c, 0xeb, 0x53,
	0x0a, 0xc6, 0x17, 0x1b, 0xfb, 0x62, 0x9d, 0x74, 0x6d, 0x5f, 0x1c, 0xd5, 0x5f, 0x98, 0x8b, 0x0d,
	0xdd, 0x75, 0xad, 0x83, 0xfb, 0xba, 0x73, 0x54, 0x5f, 0x78, 0xb3, 0xdd, 0xbb, 0xd2, 0x33, 0x72,
	0xd9, 0xb5, 0x2b, 0x4d, 0xa1, 0x5d, 0xcb, 0x72, 0xb1, 0x61, 0x4f, 0xc8, 0x6f, 0xc5, 0xda, 0xd3,
	0x46, 0x2b, 0xc6, 0x9f, 0x4e, 0xea, 0x1d, 0xe8, 0x9b, 0xf1, 0xea, 0x1d, 0x70, 0xe3, 0x05, 0xb9,
	0x9c, 0x20, 0xd4, 0xe5, 0x06, 0x4a, 0xda, 0xfa, 0x9f, 0xc3, 0x0f, 0xff, 0x07, 0x00, 0x00, 0xff,
	0xff, 0x60, 0xcb, 0x4b, 0xd0, 0xf8, 0x08, 0x00, 0x00,
}