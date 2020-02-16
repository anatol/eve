// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cipherinfo.proto

package config

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

// Security Key Exchange Method
type KeyExchangeScheme int32

const (
	KeyExchangeScheme_KEA_NONE KeyExchangeScheme = 0
	KeyExchangeScheme_KEA_ECDH KeyExchangeScheme = 1
)

var KeyExchangeScheme_name = map[int32]string{
	0: "KEA_NONE",
	1: "KEA_ECDH",
}

var KeyExchangeScheme_value = map[string]int32{
	"KEA_NONE": 0,
	"KEA_ECDH": 1,
}

func (x KeyExchangeScheme) String() string {
	return proto.EnumName(KeyExchangeScheme_name, int32(x))
}

func (KeyExchangeScheme) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{0}
}

// Encryption Scheme for Cipher Payload
type EncryptionScheme int32

const (
	EncryptionScheme_SA_NONE        EncryptionScheme = 0
	EncryptionScheme_SA_AES_256_CFB EncryptionScheme = 1
)

var EncryptionScheme_name = map[int32]string{
	0: "SA_NONE",
	1: "SA_AES_256_CFB",
}

var EncryptionScheme_value = map[string]int32{
	"SA_NONE":        0,
	"SA_AES_256_CFB": 1,
}

func (x EncryptionScheme) String() string {
	return proto.EnumName(EncryptionScheme_name, int32(x))
}

func (EncryptionScheme) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{1}
}

type CipherHashAlgorithm int32

const (
	CipherHashAlgorithm_HASH_NONE           CipherHashAlgorithm = 0
	CipherHashAlgorithm_HASH_SHA256_16bytes CipherHashAlgorithm = 1
	CipherHashAlgorithm_HASH_SHA256_32bytes CipherHashAlgorithm = 2
)

var CipherHashAlgorithm_name = map[int32]string{
	0: "HASH_NONE",
	1: "HASH_SHA256_16bytes",
	2: "HASH_SHA256_32bytes",
}

var CipherHashAlgorithm_value = map[string]int32{
	"HASH_NONE":           0,
	"HASH_SHA256_16bytes": 1,
	"HASH_SHA256_32bytes": 2,
}

func (x CipherHashAlgorithm) String() string {
	return proto.EnumName(CipherHashAlgorithm_name, int32(x))
}

func (CipherHashAlgorithm) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{2}
}

// Cipher information to decrypt Sensitive Data
type CipherContext struct {
	// cipher context id, key to this structure
	ContextId string `protobuf:"bytes,1,opt,name=contextId,proto3" json:"contextId,omitempty"`
	// algorithm used to compute hash for certificates
	HashScheme CipherHashAlgorithm `protobuf:"varint,2,opt,name=hashScheme,proto3,enum=CipherHashAlgorithm" json:"hashScheme,omitempty"`
	// for key exchange scheme, like ECDH etc.
	KeyExchangeScheme KeyExchangeScheme `protobuf:"varint,3,opt,name=keyExchangeScheme,proto3,enum=KeyExchangeScheme" json:"keyExchangeScheme,omitempty"`
	// for encrypting sensitive data, like AES256 etc.
	EncryptionScheme EncryptionScheme `protobuf:"varint,4,opt,name=encryptionScheme,proto3,enum=EncryptionScheme" json:"encryptionScheme,omitempty"`
	// device public certificate hash
	DeviceCertHash []byte `protobuf:"bytes,5,opt,name=deviceCertHash,proto3" json:"deviceCertHash,omitempty"`
	// controller certificate hash
	ControllerCertHash   []byte   `protobuf:"bytes,6,opt,name=controllerCertHash,proto3" json:"controllerCertHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CipherContext) Reset()         { *m = CipherContext{} }
func (m *CipherContext) String() string { return proto.CompactTextString(m) }
func (*CipherContext) ProtoMessage()    {}
func (*CipherContext) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{0}
}

func (m *CipherContext) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherContext.Unmarshal(m, b)
}
func (m *CipherContext) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherContext.Marshal(b, m, deterministic)
}
func (m *CipherContext) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherContext.Merge(m, src)
}
func (m *CipherContext) XXX_Size() int {
	return xxx_messageInfo_CipherContext.Size(m)
}
func (m *CipherContext) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherContext.DiscardUnknown(m)
}

var xxx_messageInfo_CipherContext proto.InternalMessageInfo

func (m *CipherContext) GetContextId() string {
	if m != nil {
		return m.ContextId
	}
	return ""
}

func (m *CipherContext) GetHashScheme() CipherHashAlgorithm {
	if m != nil {
		return m.HashScheme
	}
	return CipherHashAlgorithm_HASH_NONE
}

func (m *CipherContext) GetKeyExchangeScheme() KeyExchangeScheme {
	if m != nil {
		return m.KeyExchangeScheme
	}
	return KeyExchangeScheme_KEA_NONE
}

func (m *CipherContext) GetEncryptionScheme() EncryptionScheme {
	if m != nil {
		return m.EncryptionScheme
	}
	return EncryptionScheme_SA_NONE
}

func (m *CipherContext) GetDeviceCertHash() []byte {
	if m != nil {
		return m.DeviceCertHash
	}
	return nil
}

func (m *CipherContext) GetControllerCertHash() []byte {
	if m != nil {
		return m.ControllerCertHash
	}
	return nil
}

// Encrypted sensitive data information
type CipherBlock struct {
	// cipher context id
	CipherContextId string `protobuf:"bytes,1,opt,name=cipherContextId,proto3" json:"cipherContextId,omitempty"`
	// Initial Value for Symmetric Key derivation
	InitialValue []byte `protobuf:"bytes,2,opt,name=initialValue,proto3" json:"initialValue,omitempty"`
	// encrypted sensitive data
	CipherData []byte `protobuf:"bytes,3,opt,name=cipherData,proto3" json:"cipherData,omitempty"`
	// sha256 of the plaintext sensitive data
	ClearTextSha256      []byte   `protobuf:"bytes,4,opt,name=clearTextSha256,proto3" json:"clearTextSha256,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CipherBlock) Reset()         { *m = CipherBlock{} }
func (m *CipherBlock) String() string { return proto.CompactTextString(m) }
func (*CipherBlock) ProtoMessage()    {}
func (*CipherBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{1}
}

func (m *CipherBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherBlock.Unmarshal(m, b)
}
func (m *CipherBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherBlock.Marshal(b, m, deterministic)
}
func (m *CipherBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherBlock.Merge(m, src)
}
func (m *CipherBlock) XXX_Size() int {
	return xxx_messageInfo_CipherBlock.Size(m)
}
func (m *CipherBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherBlock.DiscardUnknown(m)
}

var xxx_messageInfo_CipherBlock proto.InternalMessageInfo

func (m *CipherBlock) GetCipherContextId() string {
	if m != nil {
		return m.CipherContextId
	}
	return ""
}

func (m *CipherBlock) GetInitialValue() []byte {
	if m != nil {
		return m.InitialValue
	}
	return nil
}

func (m *CipherBlock) GetCipherData() []byte {
	if m != nil {
		return m.CipherData
	}
	return nil
}

func (m *CipherBlock) GetClearTextSha256() []byte {
	if m != nil {
		return m.ClearTextSha256
	}
	return nil
}

// This message will be filled with the
// credential details and encrypted across
//  wire for data in transit, by the controller
// for encryption
type CredentialBlock struct {
	Identity             string   `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CredentialBlock) Reset()         { *m = CredentialBlock{} }
func (m *CredentialBlock) String() string { return proto.CompactTextString(m) }
func (*CredentialBlock) ProtoMessage()    {}
func (*CredentialBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32c1c7154980027, []int{2}
}

func (m *CredentialBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CredentialBlock.Unmarshal(m, b)
}
func (m *CredentialBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CredentialBlock.Marshal(b, m, deterministic)
}
func (m *CredentialBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CredentialBlock.Merge(m, src)
}
func (m *CredentialBlock) XXX_Size() int {
	return xxx_messageInfo_CredentialBlock.Size(m)
}
func (m *CredentialBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_CredentialBlock.DiscardUnknown(m)
}

var xxx_messageInfo_CredentialBlock proto.InternalMessageInfo

func (m *CredentialBlock) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

func (m *CredentialBlock) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func init() {
	proto.RegisterEnum("KeyExchangeScheme", KeyExchangeScheme_name, KeyExchangeScheme_value)
	proto.RegisterEnum("EncryptionScheme", EncryptionScheme_name, EncryptionScheme_value)
	proto.RegisterEnum("CipherHashAlgorithm", CipherHashAlgorithm_name, CipherHashAlgorithm_value)
	proto.RegisterType((*CipherContext)(nil), "CipherContext")
	proto.RegisterType((*CipherBlock)(nil), "CipherBlock")
	proto.RegisterType((*CredentialBlock)(nil), "CredentialBlock")
}

func init() { proto.RegisterFile("cipherinfo.proto", fileDescriptor_d32c1c7154980027) }

var fileDescriptor_d32c1c7154980027 = []byte{
	// 479 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0x6f, 0x6b, 0xd3, 0x40,
	0x18, 0x37, 0x53, 0xe7, 0xfa, 0x2c, 0xeb, 0xd2, 0x9b, 0x60, 0x11, 0xd1, 0x52, 0x44, 0x4a, 0xc1,
	0x04, 0x5b, 0xd7, 0x77, 0x82, 0x69, 0x16, 0xed, 0x18, 0x4c, 0x48, 0x64, 0x2f, 0x7c, 0x53, 0xae,
	0x97, 0xa7, 0xc9, 0xb1, 0x34, 0x57, 0x2e, 0xd7, 0xd9, 0xfa, 0x7d, 0xfc, 0x16, 0x7e, 0x38, 0xc9,
	0xa5, 0x8b, 0x35, 0xdd, 0xbb, 0x3c, 0xbf, 0x3f, 0x4f, 0x7e, 0xf9, 0xdd, 0x05, 0x2c, 0xc6, 0x97,
	0x09, 0x4a, 0x9e, 0xcd, 0x85, 0xbd, 0x94, 0x42, 0x89, 0xee, 0x9f, 0x03, 0x38, 0xf1, 0x34, 0xe8,
	0x89, 0x4c, 0xe1, 0x5a, 0x91, 0x57, 0xd0, 0x60, 0xe5, 0xe3, 0x65, 0xd4, 0x36, 0x3a, 0x46, 0xaf,
	0x11, 0xfc, 0x03, 0xc8, 0x47, 0x80, 0x84, 0xe6, 0x49, 0xc8, 0x12, 0x5c, 0x60, 0xfb, 0xa0, 0x63,
	0xf4, 0x9a, 0x83, 0xe7, 0x76, 0xb9, 0x76, 0x42, 0xf3, 0xc4, 0x4d, 0x63, 0x21, 0xb9, 0x4a, 0x16,
	0xc1, 0x8e, 0x8e, 0x7c, 0x86, 0xd6, 0x2d, 0x6e, 0xfc, 0x35, 0x4b, 0x68, 0x16, 0xe3, 0xd6, 0xfc,
	0x58, 0x9b, 0x89, 0x7d, 0x55, 0x67, 0x82, 0x7d, 0x31, 0xf9, 0x04, 0x16, 0x66, 0x4c, 0x6e, 0x96,
	0x8a, 0x8b, 0x6c, 0xbb, 0xe0, 0x89, 0x5e, 0xd0, 0xb2, 0xfd, 0x1a, 0x11, 0xec, 0x49, 0xc9, 0x3b,
	0x68, 0x46, 0x78, 0xc7, 0x19, 0x7a, 0x28, 0x55, 0x91, 0xb3, 0xfd, 0xb4, 0x63, 0xf4, 0xcc, 0xa0,
	0x86, 0x12, 0x1b, 0x48, 0xf1, 0xad, 0x52, 0xa4, 0x29, 0xca, 0x4a, 0x7b, 0xa8, 0xb5, 0x0f, 0x30,
	0xdd, 0xdf, 0x06, 0x1c, 0x97, 0xf5, 0x8d, 0x53, 0xc1, 0x6e, 0x49, 0x0f, 0x4e, 0xd9, 0x6e, 0x9b,
	0x55, 0x85, 0x75, 0x98, 0x74, 0xc1, 0xe4, 0x19, 0x57, 0x9c, 0xa6, 0x37, 0x34, 0x5d, 0x95, 0x55,
	0x9a, 0xc1, 0x7f, 0x18, 0x79, 0x0d, 0x50, 0xda, 0x2e, 0xa8, 0xa2, 0xba, 0x2f, 0x33, 0xd8, 0x41,
	0xf4, 0xdb, 0x52, 0xa4, 0xf2, 0x3b, 0xae, 0x55, 0x98, 0xd0, 0xc1, 0xf9, 0x48, 0x77, 0x62, 0x06,
	0x75, 0xb8, 0x7b, 0x09, 0xa7, 0x9e, 0xc4, 0x08, 0xb3, 0x62, 0x79, 0x19, 0xf5, 0x25, 0x1c, 0x71,
	0x0d, 0xa8, 0xcd, 0x36, 0x63, 0x35, 0x17, 0xdc, 0x92, 0xe6, 0xf9, 0x4f, 0x21, 0x23, 0x1d, 0xac,
	0x11, 0x54, 0x73, 0xdf, 0x81, 0xd6, 0xde, 0x89, 0x11, 0x13, 0x8e, 0xae, 0x7c, 0x77, 0x7a, 0xfd,
	0xed, 0xda, 0xb7, 0x1e, 0xdd, 0x4f, 0xbe, 0x77, 0x31, 0xb1, 0x8c, 0xfe, 0x10, 0xac, 0xfa, 0x09,
	0x91, 0x63, 0x78, 0x16, 0x56, 0x72, 0x02, 0xcd, 0xd0, 0x9d, 0xba, 0x7e, 0x38, 0x1d, 0x9c, 0x8f,
	0xa6, 0xde, 0x97, 0xb1, 0x65, 0xf4, 0x6f, 0xe0, 0xec, 0x81, 0x4b, 0x45, 0x4e, 0xa0, 0x31, 0x71,
	0xc3, 0xc9, 0xbd, 0xf3, 0x05, 0x9c, 0xe9, 0x31, 0x9c, 0xb8, 0x85, 0xf5, 0xc3, 0x68, 0xb6, 0x51,
	0x98, 0x5b, 0x46, 0x9d, 0x18, 0x0e, 0x4a, 0xe2, 0x60, 0xfc, 0x15, 0xde, 0x30, 0xb1, 0xb0, 0x7f,
	0x61, 0x84, 0x11, 0xb5, 0x59, 0x2a, 0x56, 0x91, 0xbd, 0xca, 0x51, 0x16, 0xb7, 0xa0, 0xfc, 0x25,
	0x7e, 0xbc, 0x8d, 0xb9, 0x4a, 0x56, 0x33, 0x9b, 0x89, 0x85, 0x93, 0xce, 0xdf, 0x63, 0x14, 0xa3,
	0x83, 0x77, 0xe8, 0xd0, 0x25, 0x77, 0x62, 0xe1, 0x30, 0x91, 0xcd, 0x79, 0x3c, 0x3b, 0xd4, 0xe2,
	0xe1, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xba, 0x1e, 0x21, 0x17, 0x53, 0x03, 0x00, 0x00,
}
