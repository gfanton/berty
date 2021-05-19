// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: errcode.proto

package errcode

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = proto.Marshal
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ErrCode int32

const (
	Undefined                                  ErrCode = 0
	ErrNotOperational                          ErrCode = 555
	TODO                                       ErrCode = 666
	ErrNotImplemented                          ErrCode = 777
	ErrInternal                                ErrCode = 888
	ErrInvalidInput                            ErrCode = 100
	ErrInvalidRange                            ErrCode = 101
	ErrMissingInput                            ErrCode = 102
	ErrSerialization                           ErrCode = 103
	ErrDeserialization                         ErrCode = 104
	ErrStreamRead                              ErrCode = 105
	ErrStreamWrite                             ErrCode = 106
	ErrStreamTransform                         ErrCode = 110
	ErrStreamSendAndClose                      ErrCode = 111
	ErrStreamHeaderWrite                       ErrCode = 112
	ErrStreamHeaderRead                        ErrCode = 115
	ErrStreamSink                              ErrCode = 113
	ErrStreamCloseAndRecv                      ErrCode = 114
	ErrMissingMapKey                           ErrCode = 107
	ErrDBWrite                                 ErrCode = 108
	ErrDBRead                                  ErrCode = 109
	ErrDBDestroy                               ErrCode = 120
	ErrDBMigrate                               ErrCode = 121
	ErrDBReplay                                ErrCode = 122
	ErrDBRestore                               ErrCode = 123
	ErrCryptoRandomGeneration                  ErrCode = 200
	ErrCryptoKeyGeneration                     ErrCode = 201
	ErrCryptoNonceGeneration                   ErrCode = 202
	ErrCryptoSignature                         ErrCode = 203
	ErrCryptoSignatureVerification             ErrCode = 204
	ErrCryptoDecrypt                           ErrCode = 205
	ErrCryptoDecryptPayload                    ErrCode = 206
	ErrCryptoEncrypt                           ErrCode = 207
	ErrCryptoKeyConversion                     ErrCode = 208
	ErrCryptoCipherInit                        ErrCode = 209
	ErrCryptoKeyDerivation                     ErrCode = 210
	ErrMap                                     ErrCode = 300
	ErrForEach                                 ErrCode = 301
	ErrKeystoreGet                             ErrCode = 400
	ErrKeystorePut                             ErrCode = 401
	ErrNotFound                                ErrCode = 404
	ErrOrbitDBInit                             ErrCode = 1000
	ErrOrbitDBOpen                             ErrCode = 1001
	ErrOrbitDBAppend                           ErrCode = 1002
	ErrOrbitDBDeserialization                  ErrCode = 1003
	ErrOrbitDBStoreCast                        ErrCode = 1004
	ErrIPFSAdd                                 ErrCode = 1050
	ErrIPFSGet                                 ErrCode = 1051
	ErrIPFSInit                                ErrCode = 1052
	ErrIPFSSetupConfig                         ErrCode = 1053
	ErrIPFSSetupRepo                           ErrCode = 1054
	ErrIPFSSetupHost                           ErrCode = 1055
	ErrHandshakeOwnEphemeralKeyGenSend         ErrCode = 1100
	ErrHandshakePeerEphemeralKeyRecv           ErrCode = 1101
	ErrHandshakeRequesterAuthenticateBoxKeyGen ErrCode = 1102
	ErrHandshakeResponderAcceptBoxKeyGen       ErrCode = 1103
	ErrHandshakeRequesterHello                 ErrCode = 1104
	ErrHandshakeResponderHello                 ErrCode = 1105
	ErrHandshakeRequesterAuthenticate          ErrCode = 1106
	ErrHandshakeResponderAccept                ErrCode = 1107
	ErrHandshakeRequesterAcknowledge           ErrCode = 1108
	ErrContactRequestSameAccount               ErrCode = 1200
	ErrContactRequestContactAlreadyAdded       ErrCode = 1201
	ErrContactRequestContactBlocked            ErrCode = 1202
	ErrContactRequestContactUndefined          ErrCode = 1203
	ErrContactRequestIncomingAlreadyReceived   ErrCode = 1204
	ErrGroupMemberLogEventOpen                 ErrCode = 1300
	ErrGroupMemberLogEventSignature            ErrCode = 1301
	ErrGroupMemberUnknownGroupID               ErrCode = 1302
	ErrGroupSecretOtherDestMember              ErrCode = 1303
	ErrGroupSecretAlreadySentToMember          ErrCode = 1304
	ErrGroupInvalidType                        ErrCode = 1305
	ErrGroupMissing                            ErrCode = 1306
	ErrGroupActivate                           ErrCode = 1307
	ErrGroupDeactivate                         ErrCode = 1308
	ErrGroupInfo                               ErrCode = 1309
	// Event errors
	ErrEventListMetadata                   ErrCode = 1400
	ErrEventListMessage                    ErrCode = 1401
	ErrMessageKeyPersistencePut            ErrCode = 1500
	ErrMessageKeyPersistenceGet            ErrCode = 1501
	ErrBridgeInterrupted                   ErrCode = 1600
	ErrBridgeNotRunning                    ErrCode = 1601
	ErrMessengerInvalidDeepLink            ErrCode = 2000
	ErrMessengerDeepLinkRequiresPassphrase ErrCode = 2001
	ErrMessengerDeepLinkInvalidPassphrase  ErrCode = 2002
	ErrMessengerStreamEvent                ErrCode = 2003
	ErrDBEntryAlreadyExists                ErrCode = 2100
	ErrDBAddConversation                   ErrCode = 2101
	ErrDBAddContactRequestOutgoingSent     ErrCode = 2102
	ErrDBAddContactRequestOutgoingEnqueud  ErrCode = 2103
	ErrDBAddContactRequestIncomingReceived ErrCode = 2104
	ErrDBAddContactRequestIncomingAccepted ErrCode = 2105
	ErrDBAddGroupMemberDeviceAdded         ErrCode = 2106
	ErrDBMultipleRecords                   ErrCode = 2107
	ErrReplayProcessGroupMetadata          ErrCode = 2200
	ErrReplayProcessGroupMessage           ErrCode = 2201
	ErrAttachmentPrepare                   ErrCode = 2300
	ErrAttachmentRetrieve                  ErrCode = 2301
	ErrProtocolSend                        ErrCode = 2302
	// Test Error
	ErrTestEcho                          ErrCode = 2401
	ErrTestEchoRecv                      ErrCode = 2402
	ErrTestEchoSend                      ErrCode = 2403
	ErrCLINoTermcaps                     ErrCode = 3001
	ErrServicesAuth                      ErrCode = 4000
	ErrServicesAuthNotInitialized        ErrCode = 4001
	ErrServicesAuthWrongState            ErrCode = 4002
	ErrServicesAuthInvalidResponse       ErrCode = 4003
	ErrServicesAuthServer                ErrCode = 4004
	ErrServicesAuthCodeChallenge         ErrCode = 4005
	ErrServicesAuthServiceInvalidToken   ErrCode = 4006
	ErrServicesAuthServiceNotSupported   ErrCode = 4007
	ErrServicesAuthUnknownToken          ErrCode = 4008
	ErrServicesAuthInvalidURL            ErrCode = 4009
	ErrServiceReplication                ErrCode = 4100
	ErrServiceReplicationServer          ErrCode = 4101
	ErrServiceReplicationMissingEndpoint ErrCode = 4102
	ErrBertyAccount                      ErrCode = 5000
	ErrBertyAccountNoIDSpecified         ErrCode = 5001
	ErrBertyAccountAlreadyOpened         ErrCode = 5002
	ErrBertyAccountInvalidIDFormat       ErrCode = 5003
	ErrBertyAccountLoggerDecorator       ErrCode = 5004
	ErrBertyAccountGRPCClient            ErrCode = 5005
	ErrBertyAccountOpenAccount           ErrCode = 5006
	ErrBertyAccountDataNotFound          ErrCode = 5007
	ErrBertyAccountMetadataUpdate        ErrCode = 5008
	ErrBertyAccountManagerOpen           ErrCode = 5009
	ErrBertyAccountManagerClose          ErrCode = 5010
	ErrBertyAccountInvalidCLIArgs        ErrCode = 5011
	ErrBertyAccountFSError               ErrCode = 5012
	ErrBertyAccountAlreadyExists         ErrCode = 5013
	ErrBertyAccountNoBackupSpecified     ErrCode = 5014
	ErrBertyAccountIDGenFailed           ErrCode = 5015
	ErrBertyAccountCreationFailed        ErrCode = 5016
	ErrBertyAccountUpdateFailed          ErrCode = 5017
)

var ErrCode_name = map[int32]string{
	0:    "Undefined",
	555:  "ErrNotOperational",
	666:  "TODO",
	777:  "ErrNotImplemented",
	888:  "ErrInternal",
	100:  "ErrInvalidInput",
	101:  "ErrInvalidRange",
	102:  "ErrMissingInput",
	103:  "ErrSerialization",
	104:  "ErrDeserialization",
	105:  "ErrStreamRead",
	106:  "ErrStreamWrite",
	110:  "ErrStreamTransform",
	111:  "ErrStreamSendAndClose",
	112:  "ErrStreamHeaderWrite",
	115:  "ErrStreamHeaderRead",
	113:  "ErrStreamSink",
	114:  "ErrStreamCloseAndRecv",
	107:  "ErrMissingMapKey",
	108:  "ErrDBWrite",
	109:  "ErrDBRead",
	120:  "ErrDBDestroy",
	121:  "ErrDBMigrate",
	122:  "ErrDBReplay",
	123:  "ErrDBRestore",
	200:  "ErrCryptoRandomGeneration",
	201:  "ErrCryptoKeyGeneration",
	202:  "ErrCryptoNonceGeneration",
	203:  "ErrCryptoSignature",
	204:  "ErrCryptoSignatureVerification",
	205:  "ErrCryptoDecrypt",
	206:  "ErrCryptoDecryptPayload",
	207:  "ErrCryptoEncrypt",
	208:  "ErrCryptoKeyConversion",
	209:  "ErrCryptoCipherInit",
	210:  "ErrCryptoKeyDerivation",
	300:  "ErrMap",
	301:  "ErrForEach",
	400:  "ErrKeystoreGet",
	401:  "ErrKeystorePut",
	404:  "ErrNotFound",
	1000: "ErrOrbitDBInit",
	1001: "ErrOrbitDBOpen",
	1002: "ErrOrbitDBAppend",
	1003: "ErrOrbitDBDeserialization",
	1004: "ErrOrbitDBStoreCast",
	1050: "ErrIPFSAdd",
	1051: "ErrIPFSGet",
	1052: "ErrIPFSInit",
	1053: "ErrIPFSSetupConfig",
	1054: "ErrIPFSSetupRepo",
	1055: "ErrIPFSSetupHost",
	1100: "ErrHandshakeOwnEphemeralKeyGenSend",
	1101: "ErrHandshakePeerEphemeralKeyRecv",
	1102: "ErrHandshakeRequesterAuthenticateBoxKeyGen",
	1103: "ErrHandshakeResponderAcceptBoxKeyGen",
	1104: "ErrHandshakeRequesterHello",
	1105: "ErrHandshakeResponderHello",
	1106: "ErrHandshakeRequesterAuthenticate",
	1107: "ErrHandshakeResponderAccept",
	1108: "ErrHandshakeRequesterAcknowledge",
	1200: "ErrContactRequestSameAccount",
	1201: "ErrContactRequestContactAlreadyAdded",
	1202: "ErrContactRequestContactBlocked",
	1203: "ErrContactRequestContactUndefined",
	1204: "ErrContactRequestIncomingAlreadyReceived",
	1300: "ErrGroupMemberLogEventOpen",
	1301: "ErrGroupMemberLogEventSignature",
	1302: "ErrGroupMemberUnknownGroupID",
	1303: "ErrGroupSecretOtherDestMember",
	1304: "ErrGroupSecretAlreadySentToMember",
	1305: "ErrGroupInvalidType",
	1306: "ErrGroupMissing",
	1307: "ErrGroupActivate",
	1308: "ErrGroupDeactivate",
	1309: "ErrGroupInfo",
	1400: "ErrEventListMetadata",
	1401: "ErrEventListMessage",
	1500: "ErrMessageKeyPersistencePut",
	1501: "ErrMessageKeyPersistenceGet",
	1600: "ErrBridgeInterrupted",
	1601: "ErrBridgeNotRunning",
	2000: "ErrMessengerInvalidDeepLink",
	2001: "ErrMessengerDeepLinkRequiresPassphrase",
	2002: "ErrMessengerDeepLinkInvalidPassphrase",
	2003: "ErrMessengerStreamEvent",
	2100: "ErrDBEntryAlreadyExists",
	2101: "ErrDBAddConversation",
	2102: "ErrDBAddContactRequestOutgoingSent",
	2103: "ErrDBAddContactRequestOutgoingEnqueud",
	2104: "ErrDBAddContactRequestIncomingReceived",
	2105: "ErrDBAddContactRequestIncomingAccepted",
	2106: "ErrDBAddGroupMemberDeviceAdded",
	2107: "ErrDBMultipleRecords",
	2200: "ErrReplayProcessGroupMetadata",
	2201: "ErrReplayProcessGroupMessage",
	2300: "ErrAttachmentPrepare",
	2301: "ErrAttachmentRetrieve",
	2302: "ErrProtocolSend",
	2401: "ErrTestEcho",
	2402: "ErrTestEchoRecv",
	2403: "ErrTestEchoSend",
	3001: "ErrCLINoTermcaps",
	4000: "ErrServicesAuth",
	4001: "ErrServicesAuthNotInitialized",
	4002: "ErrServicesAuthWrongState",
	4003: "ErrServicesAuthInvalidResponse",
	4004: "ErrServicesAuthServer",
	4005: "ErrServicesAuthCodeChallenge",
	4006: "ErrServicesAuthServiceInvalidToken",
	4007: "ErrServicesAuthServiceNotSupported",
	4008: "ErrServicesAuthUnknownToken",
	4009: "ErrServicesAuthInvalidURL",
	4100: "ErrServiceReplication",
	4101: "ErrServiceReplicationServer",
	4102: "ErrServiceReplicationMissingEndpoint",
	5000: "ErrBertyAccount",
	5001: "ErrBertyAccountNoIDSpecified",
	5002: "ErrBertyAccountAlreadyOpened",
	5003: "ErrBertyAccountInvalidIDFormat",
	5004: "ErrBertyAccountLoggerDecorator",
	5005: "ErrBertyAccountGRPCClient",
	5006: "ErrBertyAccountOpenAccount",
	5007: "ErrBertyAccountDataNotFound",
	5008: "ErrBertyAccountMetadataUpdate",
	5009: "ErrBertyAccountManagerOpen",
	5010: "ErrBertyAccountManagerClose",
	5011: "ErrBertyAccountInvalidCLIArgs",
	5012: "ErrBertyAccountFSError",
	5013: "ErrBertyAccountAlreadyExists",
	5014: "ErrBertyAccountNoBackupSpecified",
	5015: "ErrBertyAccountIDGenFailed",
	5016: "ErrBertyAccountCreationFailed",
	5017: "ErrBertyAccountUpdateFailed",
}

var ErrCode_value = map[string]int32{
	"Undefined":                          0,
	"ErrNotOperational":                  555,
	"TODO":                               666,
	"ErrNotImplemented":                  777,
	"ErrInternal":                        888,
	"ErrInvalidInput":                    100,
	"ErrInvalidRange":                    101,
	"ErrMissingInput":                    102,
	"ErrSerialization":                   103,
	"ErrDeserialization":                 104,
	"ErrStreamRead":                      105,
	"ErrStreamWrite":                     106,
	"ErrStreamTransform":                 110,
	"ErrStreamSendAndClose":              111,
	"ErrStreamHeaderWrite":               112,
	"ErrStreamHeaderRead":                115,
	"ErrStreamSink":                      113,
	"ErrStreamCloseAndRecv":              114,
	"ErrMissingMapKey":                   107,
	"ErrDBWrite":                         108,
	"ErrDBRead":                          109,
	"ErrDBDestroy":                       120,
	"ErrDBMigrate":                       121,
	"ErrDBReplay":                        122,
	"ErrDBRestore":                       123,
	"ErrCryptoRandomGeneration":          200,
	"ErrCryptoKeyGeneration":             201,
	"ErrCryptoNonceGeneration":           202,
	"ErrCryptoSignature":                 203,
	"ErrCryptoSignatureVerification":     204,
	"ErrCryptoDecrypt":                   205,
	"ErrCryptoDecryptPayload":            206,
	"ErrCryptoEncrypt":                   207,
	"ErrCryptoKeyConversion":             208,
	"ErrCryptoCipherInit":                209,
	"ErrCryptoKeyDerivation":             210,
	"ErrMap":                             300,
	"ErrForEach":                         301,
	"ErrKeystoreGet":                     400,
	"ErrKeystorePut":                     401,
	"ErrNotFound":                        404,
	"ErrOrbitDBInit":                     1000,
	"ErrOrbitDBOpen":                     1001,
	"ErrOrbitDBAppend":                   1002,
	"ErrOrbitDBDeserialization":          1003,
	"ErrOrbitDBStoreCast":                1004,
	"ErrIPFSAdd":                         1050,
	"ErrIPFSGet":                         1051,
	"ErrIPFSInit":                        1052,
	"ErrIPFSSetupConfig":                 1053,
	"ErrIPFSSetupRepo":                   1054,
	"ErrIPFSSetupHost":                   1055,
	"ErrHandshakeOwnEphemeralKeyGenSend": 1100,
	"ErrHandshakePeerEphemeralKeyRecv":   1101,
	"ErrHandshakeRequesterAuthenticateBoxKeyGen": 1102,
	"ErrHandshakeResponderAcceptBoxKeyGen":       1103,
	"ErrHandshakeRequesterHello":                 1104,
	"ErrHandshakeResponderHello":                 1105,
	"ErrHandshakeRequesterAuthenticate":          1106,
	"ErrHandshakeResponderAccept":                1107,
	"ErrHandshakeRequesterAcknowledge":           1108,
	"ErrContactRequestSameAccount":               1200,
	"ErrContactRequestContactAlreadyAdded":       1201,
	"ErrContactRequestContactBlocked":            1202,
	"ErrContactRequestContactUndefined":          1203,
	"ErrContactRequestIncomingAlreadyReceived":   1204,
	"ErrGroupMemberLogEventOpen":                 1300,
	"ErrGroupMemberLogEventSignature":            1301,
	"ErrGroupMemberUnknownGroupID":               1302,
	"ErrGroupSecretOtherDestMember":              1303,
	"ErrGroupSecretAlreadySentToMember":          1304,
	"ErrGroupInvalidType":                        1305,
	"ErrGroupMissing":                            1306,
	"ErrGroupActivate":                           1307,
	"ErrGroupDeactivate":                         1308,
	"ErrGroupInfo":                               1309,
	"ErrEventListMetadata":                       1400,
	"ErrEventListMessage":                        1401,
	"ErrMessageKeyPersistencePut":                1500,
	"ErrMessageKeyPersistenceGet":                1501,
	"ErrBridgeInterrupted":                       1600,
	"ErrBridgeNotRunning":                        1601,
	"ErrMessengerInvalidDeepLink":                2000,
	"ErrMessengerDeepLinkRequiresPassphrase":     2001,
	"ErrMessengerDeepLinkInvalidPassphrase":      2002,
	"ErrMessengerStreamEvent":                    2003,
	"ErrDBEntryAlreadyExists":                    2100,
	"ErrDBAddConversation":                       2101,
	"ErrDBAddContactRequestOutgoingSent":         2102,
	"ErrDBAddContactRequestOutgoingEnqueud":      2103,
	"ErrDBAddContactRequestIncomingReceived":     2104,
	"ErrDBAddContactRequestIncomingAccepted":     2105,
	"ErrDBAddGroupMemberDeviceAdded":             2106,
	"ErrDBMultipleRecords":                       2107,
	"ErrReplayProcessGroupMetadata":              2200,
	"ErrReplayProcessGroupMessage":               2201,
	"ErrAttachmentPrepare":                       2300,
	"ErrAttachmentRetrieve":                      2301,
	"ErrProtocolSend":                            2302,
	"ErrTestEcho":                                2401,
	"ErrTestEchoRecv":                            2402,
	"ErrTestEchoSend":                            2403,
	"ErrCLINoTermcaps":                           3001,
	"ErrServicesAuth":                            4000,
	"ErrServicesAuthNotInitialized":              4001,
	"ErrServicesAuthWrongState":                  4002,
	"ErrServicesAuthInvalidResponse":             4003,
	"ErrServicesAuthServer":                      4004,
	"ErrServicesAuthCodeChallenge":               4005,
	"ErrServicesAuthServiceInvalidToken":         4006,
	"ErrServicesAuthServiceNotSupported":         4007,
	"ErrServicesAuthUnknownToken":                4008,
	"ErrServicesAuthInvalidURL":                  4009,
	"ErrServiceReplication":                      4100,
	"ErrServiceReplicationServer":                4101,
	"ErrServiceReplicationMissingEndpoint":       4102,
	"ErrBertyAccount":                            5000,
	"ErrBertyAccountNoIDSpecified":               5001,
	"ErrBertyAccountAlreadyOpened":               5002,
	"ErrBertyAccountInvalidIDFormat":             5003,
	"ErrBertyAccountLoggerDecorator":             5004,
	"ErrBertyAccountGRPCClient":                  5005,
	"ErrBertyAccountOpenAccount":                 5006,
	"ErrBertyAccountDataNotFound":                5007,
	"ErrBertyAccountMetadataUpdate":              5008,
	"ErrBertyAccountManagerOpen":                 5009,
	"ErrBertyAccountManagerClose":                5010,
	"ErrBertyAccountInvalidCLIArgs":              5011,
	"ErrBertyAccountFSError":                     5012,
	"ErrBertyAccountAlreadyExists":               5013,
	"ErrBertyAccountNoBackupSpecified":           5014,
	"ErrBertyAccountIDGenFailed":                 5015,
	"ErrBertyAccountCreationFailed":              5016,
	"ErrBertyAccountUpdateFailed":                5017,
}

func (x ErrCode) String() string {
	return proto.EnumName(ErrCode_name, int32(x))
}

func (ErrCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_4240057316120df7, []int{0}
}

type ErrDetails struct {
	Codes                []ErrCode `protobuf:"varint,1,rep,packed,name=codes,proto3,enum=berty.errcode.ErrCode" json:"codes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ErrDetails) Reset()         { *m = ErrDetails{} }
func (m *ErrDetails) String() string { return proto.CompactTextString(m) }
func (*ErrDetails) ProtoMessage()    {}
func (*ErrDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_4240057316120df7, []int{0}
}

func (m *ErrDetails) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *ErrDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ErrDetails.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *ErrDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrDetails.Merge(m, src)
}

func (m *ErrDetails) XXX_Size() int {
	return m.Size()
}

func (m *ErrDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrDetails.DiscardUnknown(m)
}

var xxx_messageInfo_ErrDetails proto.InternalMessageInfo

func (m *ErrDetails) GetCodes() []ErrCode {
	if m != nil {
		return m.Codes
	}
	return nil
}

func init() {
	proto.RegisterEnum("berty.errcode.ErrCode", ErrCode_name, ErrCode_value)
	proto.RegisterType((*ErrDetails)(nil), "berty.errcode.ErrDetails")
}

func init() { proto.RegisterFile("errcode.proto", fileDescriptor_4240057316120df7) }

var fileDescriptor_4240057316120df7 = []byte{
	// 1846 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x57, 0x59, 0x93, 0x5b, 0x47,
	0x15, 0x8e, 0xe2, 0xc4, 0xa3, 0x69, 0x2f, 0x73, 0xa6, 0xed, 0x78, 0x4b, 0x32, 0x9a, 0x98, 0x24,
	0x0a, 0x06, 0x3c, 0x05, 0xbc, 0xf1, 0xa6, 0x6d, 0x6c, 0x95, 0x67, 0x51, 0x49, 0x63, 0x52, 0xc5,
	0x5b, 0xfb, 0xf6, 0x99, 0xab, 0xcb, 0x5c, 0x75, 0x5f, 0xf7, 0x6d, 0x4d, 0xac, 0xf0, 0x0a, 0x14,
	0x3b, 0x09, 0x38, 0x89, 0xed, 0x24, 0xec, 0x6b, 0x01, 0x55, 0x2c, 0x61, 0x09, 0xbc, 0xc0, 0x5b,
	0x80, 0x2c, 0x63, 0x87, 0x37, 0xa0, 0x8a, 0x98, 0x17, 0xb6, 0x1f, 0x10, 0xaa, 0x80, 0xa2, 0xba,
	0xef, 0x91, 0xe6, 0x6a, 0x46, 0x0c, 0x79, 0x92, 0xee, 0x39, 0x5f, 0x9f, 0x3e, 0xfb, 0x39, 0xcd,
	0x0e, 0xa1, 0x31, 0x81, 0x96, 0x78, 0x36, 0x31, 0xda, 0x6a, 0x7e, 0xe8, 0x12, 0x1a, 0x3b, 0x38,
	0x4b, 0xc4, 0x53, 0x47, 0x43, 0x1d, 0x6a, 0xcf, 0x59, 0x70, 0xff, 0x32, 0xd0, 0xe9, 0xf7, 0x31,
	0xd6, 0x30, 0xa6, 0x8e, 0x56, 0x44, 0x71, 0xca, 0xdf, 0xc9, 0xee, 0x76, 0xd8, 0xf4, 0x44, 0x61,
	0x7e, 0xdf, 0x23, 0x87, 0xdf, 0x73, 0xec, 0xec, 0x98, 0x88, 0xb3, 0x0d, 0x63, 0x6a, 0x5a, 0x62,
	0x3b, 0x03, 0x9d, 0xf9, 0xfd, 0x1c, 0x9b, 0x22, 0x12, 0x3f, 0xc4, 0xa6, 0x2f, 0x2a, 0x89, 0xeb,
	0x91, 0x42, 0x09, 0x77, 0xf0, 0x63, 0x6c, 0xb6, 0x61, 0xcc, 0x8a, 0xb6, 0xab, 0x09, 0x1a, 0x61,
	0x23, 0xad, 0x44, 0x0c, 0xdf, 0xb9, 0x8b, 0x4f, 0xb3, 0xbb, 0xd6, 0x56, 0xeb, 0xab, 0x70, 0xe3,
	0xee, 0x6d, 0x48, 0xb3, 0x97, 0xc4, 0xd8, 0x43, 0x65, 0x51, 0xc2, 0xc7, 0xf7, 0x73, 0x60, 0x07,
	0x1a, 0xc6, 0x34, 0x95, 0x45, 0xe3, 0x0e, 0xbd, 0xb9, 0x9f, 0x1f, 0x61, 0x33, 0x9e, 0xb2, 0x29,
	0xe2, 0x48, 0x36, 0x55, 0xd2, 0xb7, 0x20, 0xc7, 0x89, 0x6d, 0xa1, 0x42, 0x04, 0x24, 0xe2, 0x72,
	0x94, 0xa6, 0x91, 0x0a, 0x33, 0xe4, 0x3a, 0x3f, 0xca, 0xa0, 0x61, 0x4c, 0x07, 0x4d, 0x24, 0xe2,
	0xe8, 0x71, 0xaf, 0x0c, 0x84, 0xfc, 0x18, 0xe3, 0xde, 0xf0, 0x74, 0x8c, 0xde, 0xe5, 0xb3, 0xec,
	0x90, 0x43, 0x5b, 0x83, 0xa2, 0xd7, 0x46, 0x21, 0x21, 0xe2, 0x9c, 0x1d, 0x1e, 0x91, 0x1e, 0x35,
	0x91, 0x45, 0xf8, 0x20, 0x1d, 0xcf, 0x68, 0x6b, 0x46, 0xa8, 0x74, 0x5d, 0x9b, 0x1e, 0x28, 0x7e,
	0x92, 0xdd, 0x33, 0xa2, 0x77, 0x50, 0xc9, 0x8a, 0x92, 0xb5, 0x58, 0xa7, 0x08, 0x9a, 0x9f, 0x60,
	0x47, 0x47, 0xac, 0xf3, 0x28, 0x24, 0x9a, 0x4c, 0x58, 0xc2, 0x8f, 0xb3, 0x23, 0x3b, 0x38, 0xfe,
	0xe6, 0x74, 0x4c, 0x99, 0x4e, 0xa4, 0x36, 0xe0, 0xf2, 0xd8, 0x05, 0x5e, 0x72, 0x45, 0xc9, 0x36,
	0x06, 0x9b, 0x60, 0xc8, 0x50, 0xb2, 0x7e, 0x59, 0x24, 0x17, 0x70, 0x00, 0x1b, 0xfc, 0x70, 0x16,
	0xe1, 0x6a, 0x76, 0x59, 0xec, 0x22, 0xe5, 0xbf, 0xfd, 0x15, 0x3d, 0x0e, 0xec, 0xa0, 0xff, 0xac,
	0x63, 0x6a, 0x8d, 0x1e, 0xc0, 0x95, 0x11, 0x65, 0x39, 0x0a, 0x8d, 0xb0, 0x08, 0x03, 0x3e, 0xe3,
	0x43, 0xe2, 0x8e, 0x24, 0xb1, 0x18, 0xc0, 0xe3, 0x23, 0x48, 0x1b, 0x53, 0xab, 0x0d, 0xc2, 0x87,
	0xf8, 0x1c, 0x3b, 0xe9, 0x52, 0xc1, 0x0c, 0x12, 0xab, 0xdb, 0x42, 0x49, 0xdd, 0x3b, 0x87, 0x8a,
	0x42, 0x0f, 0x2f, 0x15, 0xf8, 0xbd, 0xec, 0xd8, 0x88, 0x7f, 0x01, 0x07, 0x39, 0xe6, 0xaf, 0x0b,
	0xfc, 0x7e, 0x76, 0x62, 0xc4, 0x5c, 0xd1, 0x2a, 0xc0, 0x1c, 0xfb, 0x37, 0x05, 0x7e, 0xdc, 0xfb,
	0x3a, 0x63, 0x77, 0xa2, 0x50, 0x09, 0xdb, 0x37, 0x08, 0xbf, 0x2d, 0xf0, 0xb7, 0xb1, 0xb9, 0xdd,
	0x8c, 0xf7, 0xa3, 0x89, 0xd6, 0xa3, 0x20, 0x3b, 0xfd, 0x72, 0x81, 0xdf, 0xe3, 0xbd, 0x92, 0x81,
	0xea, 0x18, 0xb8, 0x5f, 0x78, 0xa5, 0xc0, 0xef, 0x63, 0xc7, 0x77, 0x92, 0x5b, 0x62, 0x10, 0x6b,
	0x21, 0xe1, 0xd5, 0xf1, 0x43, 0x0d, 0x95, 0x1d, 0x7a, 0x6d, 0x97, 0x15, 0x35, 0xad, 0x36, 0xd1,
	0xa4, 0xee, 0xa2, 0xad, 0x02, 0x3f, 0xe1, 0xa3, 0x98, 0x31, 0x6b, 0x51, 0xd2, 0x45, 0xd3, 0x54,
	0x91, 0x85, 0x9b, 0xbb, 0x8e, 0xd5, 0xd1, 0x44, 0x9b, 0x99, 0x7e, 0xb7, 0x0a, 0xfc, 0x00, 0xdb,
	0xef, 0xa2, 0x26, 0x12, 0xf8, 0xee, 0x9d, 0x7c, 0xc6, 0x07, 0x6b, 0x51, 0x9b, 0x86, 0x08, 0xba,
	0xf0, 0xbd, 0x3b, 0xf9, 0x11, 0x9f, 0x7b, 0x17, 0x70, 0xe0, 0x1d, 0x7d, 0x0e, 0x2d, 0x3c, 0xb1,
	0x6f, 0x07, 0xb1, 0xd5, 0xb7, 0xf0, 0xe4, 0x3e, 0xaa, 0x9b, 0x15, 0x6d, 0x17, 0x75, 0x5f, 0x49,
	0xb8, 0x3a, 0x84, 0xad, 0x9a, 0x4b, 0x91, 0xad, 0x57, 0xbd, 0x2e, 0x7f, 0x99, 0x1a, 0x27, 0xae,
	0x26, 0xa8, 0xe0, 0xaf, 0x53, 0x64, 0x2e, 0x11, 0x2b, 0x49, 0x82, 0x4a, 0xc2, 0xdf, 0xa6, 0x28,
	0xa8, 0x44, 0xde, 0x59, 0x2a, 0x7f, 0x9f, 0x22, 0x8b, 0x89, 0xdf, 0x71, 0xba, 0xd4, 0x44, 0x6a,
	0xe1, 0x1f, 0x53, 0x64, 0x47, 0xb3, 0xb5, 0xd8, 0xa9, 0x48, 0x09, 0x37, 0x8a, 0x39, 0x82, 0xb3,
	0xe1, 0xd9, 0xe2, 0xb0, 0xcc, 0x5b, 0x8b, 0x1d, 0xaf, 0xd9, 0x73, 0x45, 0x0a, 0xb3, 0xa3, 0x74,
	0xd0, 0xf6, 0x93, 0x9a, 0x56, 0xeb, 0x51, 0x08, 0xcf, 0x17, 0x49, 0xbb, 0x11, 0xa3, 0x8d, 0x89,
	0x86, 0x2f, 0xec, 0x22, 0x9f, 0xd7, 0xa9, 0x85, 0x2f, 0x16, 0x79, 0x99, 0x9d, 0x6e, 0x18, 0x73,
	0x5e, 0x28, 0x99, 0x76, 0xc5, 0x06, 0xae, 0x3e, 0xa6, 0x1a, 0x49, 0x17, 0x7b, 0x68, 0x44, 0x9c,
	0x25, 0x9e, 0x2b, 0x4b, 0x78, 0xb9, 0xc8, 0x1f, 0x62, 0xf3, 0x79, 0x60, 0x0b, 0xd1, 0xe4, 0x91,
	0xbe, 0xa8, 0x5e, 0x29, 0xf2, 0x05, 0x76, 0x26, 0x0f, 0x6b, 0xe3, 0xe5, 0x3e, 0xa6, 0x16, 0x4d,
	0xa5, 0x6f, 0xbb, 0xa8, 0xac, 0xcb, 0x34, 0xac, 0xea, 0x2b, 0x99, 0x6c, 0x78, 0xb5, 0xc8, 0xdf,
	0xce, 0x1e, 0x1c, 0x3f, 0x90, 0x26, 0x5a, 0x49, 0x34, 0x95, 0x20, 0xc0, 0xc4, 0x6e, 0x43, 0x5f,
	0x2b, 0xf2, 0x12, 0x3b, 0x35, 0x51, 0xf6, 0x79, 0x8c, 0x63, 0x0d, 0x5b, 0x13, 0x00, 0x24, 0x2b,
	0x03, 0xdc, 0x2c, 0xf2, 0x87, 0xd9, 0x03, 0xff, 0x57, 0x3b, 0xb8, 0x55, 0xe4, 0xf3, 0xec, 0xde,
	0x3d, 0x94, 0x82, 0xd7, 0x77, 0xb9, 0x63, 0x5b, 0x52, 0xb0, 0xa1, 0xf4, 0x63, 0x31, 0xca, 0x10,
	0xe1, 0x77, 0x45, 0xfe, 0x00, 0xbb, 0xcf, 0xf7, 0x7c, 0x65, 0x45, 0x60, 0x09, 0xd4, 0x11, 0x3d,
	0xac, 0x04, 0x81, 0xee, 0x2b, 0x0b, 0xdf, 0x9f, 0x26, 0x07, 0x8c, 0x43, 0xe8, 0xab, 0x12, 0x1b,
	0x14, 0x72, 0x50, 0x91, 0x12, 0x25, 0xfc, 0x60, 0x9a, 0x3f, 0xc8, 0x4a, 0xff, 0x0b, 0x5a, 0x8d,
	0x75, 0xb0, 0x81, 0x12, 0x7e, 0x38, 0x4d, 0x46, 0x4e, 0x44, 0x6d, 0x0f, 0x9d, 0x1f, 0x4d, 0xf3,
	0x77, 0xb1, 0x47, 0x76, 0xe1, 0x9a, 0x2a, 0xd0, 0xbd, 0x48, 0x85, 0x74, 0x73, 0x1b, 0x03, 0x8c,
	0x36, 0x51, 0xc2, 0x0b, 0xd3, 0xe4, 0xdc, 0x73, 0x46, 0xf7, 0x93, 0x65, 0xec, 0x5d, 0x42, 0xb3,
	0xa4, 0xc3, 0xc6, 0x26, 0x2a, 0xeb, 0xcb, 0xe2, 0x2a, 0x23, 0xed, 0x26, 0x00, 0xb6, 0xbb, 0xd0,
	0x53, 0x8c, 0x3c, 0x92, 0x43, 0x5d, 0x54, 0xce, 0x63, 0xca, 0x53, 0x9a, 0x75, 0x78, 0x9a, 0xf1,
	0xd3, 0xec, 0xfe, 0x21, 0xa4, 0x83, 0x81, 0x41, 0xbb, 0x6a, 0xbb, 0xe8, 0x86, 0x8f, 0xcd, 0x4e,
	0xc0, 0x33, 0x8c, 0x8c, 0xcc, 0x61, 0x48, 0xe3, 0x0e, 0x2a, 0xbb, 0xa6, 0x09, 0x77, 0x8d, 0x51,
	0xd1, 0x65, 0xc2, 0xb3, 0xe9, 0xb7, 0x36, 0x48, 0x10, 0xae, 0x33, 0x7e, 0xd4, 0x4f, 0xbf, 0x4c,
	0x91, 0x6c, 0x08, 0xc0, 0x0d, 0x46, 0x65, 0xe2, 0xa9, 0x95, 0xc0, 0xba, 0xc6, 0x83, 0xf0, 0x2c,
	0xa3, 0x6a, 0xf3, 0xe4, 0x3a, 0x8a, 0x21, 0xe3, 0x39, 0xc6, 0x67, 0x7d, 0x6f, 0x27, 0xf9, 0xeb,
	0x1a, 0x9e, 0x67, 0xfc, 0xa4, 0x9f, 0x5c, 0xde, 0xf2, 0xa5, 0xc8, 0xe9, 0x6c, 0x85, 0x14, 0x56,
	0xc0, 0x9b, 0x43, 0x6d, 0x72, 0xac, 0x34, 0x15, 0x21, 0xc2, 0x3f, 0x19, 0x65, 0x1c, 0x11, 0x2e,
	0xe0, 0xa0, 0xe5, 0x3a, 0x65, 0x6a, 0x51, 0x05, 0xbe, 0x63, 0xfd, 0xe1, 0xc0, 0x5e, 0x08, 0xd7,
	0x24, 0xfe, 0x78, 0x80, 0x2e, 0xae, 0x9a, 0x48, 0x86, 0xe8, 0x37, 0x02, 0xd3, 0x4f, 0xdc, 0x9a,
	0xf0, 0xcb, 0x83, 0x74, 0x71, 0xc6, 0x5a, 0xd1, 0xb6, 0xdd, 0x57, 0xca, 0x19, 0xfc, 0xab, 0x83,
	0x39, 0xb1, 0xa8, 0x42, 0x1c, 0xae, 0x08, 0x75, 0xc4, 0x64, 0xc9, 0x8d, 0xd0, 0xad, 0x19, 0xfe,
	0x0e, 0xf6, 0x70, 0x1e, 0x31, 0x64, 0xb9, 0x8c, 0x89, 0x0c, 0xa6, 0x2d, 0x91, 0xa6, 0x49, 0xd7,
	0x88, 0x14, 0xe1, 0xe6, 0x0c, 0x3f, 0xc3, 0x1e, 0x9a, 0x04, 0x26, 0xb1, 0x39, 0xec, 0xad, 0x19,
	0x1a, 0x2a, 0x23, 0x6c, 0x36, 0xa5, 0xbd, 0x73, 0xe0, 0xf5, 0x21, 0xb7, 0x5e, 0x6d, 0x28, 0x6b,
	0x06, 0x14, 0xdd, 0xc6, 0x95, 0x28, 0xb5, 0x29, 0xbc, 0x00, 0x64, 0x6b, 0xbd, 0x5a, 0x91, 0x92,
	0xe6, 0x4a, 0xd6, 0x67, 0x7f, 0x0c, 0xd4, 0xd2, 0x86, 0xac, 0x5c, 0x72, 0xaf, 0xf6, 0x6d, 0xa8,
	0x23, 0x15, 0xba, 0x1c, 0x81, 0x9f, 0x00, 0xe9, 0xba, 0x07, 0xb0, 0xa1, 0x2e, 0xf7, 0xb1, 0x2f,
	0xe1, 0xa7, 0x40, 0x4e, 0x98, 0x80, 0x1d, 0x56, 0xcc, 0xa8, 0x54, 0x7e, 0xf6, 0x16, 0xc0, 0x59,
	0x1f, 0x41, 0x09, 0x2f, 0x02, 0x8d, 0x65, 0x0f, 0xce, 0x55, 0x45, 0x1d, 0x37, 0xa3, 0x00, 0xb3,
	0xca, 0xff, 0xf9, 0xb6, 0xb9, 0xcb, 0xfd, 0xd8, 0x46, 0x49, 0x8c, 0x6d, 0x0c, 0xb4, 0x91, 0x29,
	0xfc, 0x02, 0xa8, 0x5a, 0xb2, 0x65, 0xa3, 0x65, 0x74, 0x80, 0x69, 0x4a, 0x72, 0x28, 0xef, 0xae,
	0xcd, 0x52, 0xd1, 0x4d, 0xc2, 0x64, 0x09, 0x78, 0x7d, 0x96, 0x6e, 0xa8, 0x58, 0x2b, 0x82, 0xae,
	0x5b, 0x2f, 0x5b, 0x06, 0x13, 0x61, 0x10, 0xfe, 0x35, 0xcb, 0x4f, 0xf9, 0x25, 0x6a, 0x9b, 0xd5,
	0x46, 0x6b, 0x22, 0xdc, 0x44, 0xf8, 0xf7, 0x2c, 0x55, 0x51, 0xcb, 0x6d, 0xc7, 0x81, 0x8e, 0xfd,
	0xb0, 0xf8, 0xcf, 0x2c, 0x8d, 0xab, 0x35, 0x4c, 0x6d, 0x23, 0xe8, 0x6a, 0x78, 0x83, 0x13, 0x6e,
	0x48, 0xf1, 0xd3, 0xe2, 0xf6, 0x4e, 0xaa, 0x3f, 0xfd, 0x67, 0x3e, 0x5c, 0x27, 0x96, 0x9a, 0x2b,
	0x7a, 0x0d, 0x4d, 0x2f, 0x10, 0x49, 0x0a, 0x2f, 0x1e, 0x27, 0x70, 0x07, 0x8d, 0xf3, 0x4c, 0xea,
	0x5a, 0x36, 0x7c, 0xa9, 0x44, 0xe6, 0xe7, 0xa9, 0x6e, 0x49, 0x56, 0x91, 0xf5, 0xc3, 0x17, 0x25,
	0x7c, 0xb9, 0x44, 0x93, 0x39, 0x8f, 0x79, 0xd4, 0x68, 0x15, 0x76, 0xac, 0x2b, 0xe2, 0xaf, 0x94,
	0x28, 0x04, 0x79, 0xfe, 0x70, 0x53, 0xf6, 0x8d, 0x3f, 0x45, 0xf8, 0x6a, 0x89, 0xbc, 0x90, 0x07,
	0xb9, 0xff, 0x68, 0xe0, 0x6b, 0x25, 0xf2, 0x6f, 0x9e, 0xe7, 0xd6, 0xfc, 0x5a, 0x57, 0xc4, 0xb1,
	0x4b, 0x6d, 0xf8, 0x7a, 0x89, 0xb2, 0x72, 0xe7, 0xf1, 0x28, 0xc0, 0x61, 0x5b, 0xd2, 0x1b, 0xa8,
	0xe0, 0x1b, 0x7b, 0x00, 0x57, 0xb4, 0xed, 0xf4, 0x93, 0x44, 0x1b, 0x97, 0x38, 0xdf, 0x2c, 0x51,
	0xe5, 0xe6, 0x81, 0xd4, 0x4a, 0x33, 0x51, 0xdf, 0x9a, 0x64, 0x37, 0x5d, 0x76, 0xb1, 0xbd, 0x04,
	0xdf, 0xde, 0x61, 0x92, 0xcb, 0x8e, 0xe1, 0x22, 0xf8, 0xe1, 0xf9, 0x71, 0xe9, 0x39, 0x1e, 0x19,
	0xfd, 0x91, 0x79, 0x1a, 0x5c, 0xbb, 0x11, 0xd4, 0x4d, 0x1b, 0x4a, 0x26, 0x3a, 0x52, 0x16, 0x3e,
	0x3a, 0x4f, 0xa1, 0xab, 0xba, 0xe7, 0xd1, 0x70, 0xf2, 0x7d, 0xac, 0x4c, 0x5e, 0xcb, 0x53, 0x57,
	0x74, 0xb3, 0xde, 0x49, 0x30, 0x88, 0xd6, 0x23, 0xf7, 0xbc, 0x99, 0x04, 0xa1, 0x4e, 0xe0, 0x86,
	0x0e, 0x4a, 0xf8, 0x44, 0x99, 0x82, 0x97, 0x87, 0x0c, 0xdf, 0x3e, 0xf5, 0x45, 0x6d, 0x7a, 0xc2,
	0xc2, 0x27, 0x27, 0x81, 0x96, 0x74, 0xe8, 0xfb, 0x53, 0xa0, 0x8d, 0xb0, 0xda, 0xc0, 0xa7, 0xca,
	0xe4, 0xae, 0x3c, 0xe8, 0x5c, 0xbb, 0x55, 0xab, 0xc5, 0x91, 0xeb, 0x17, 0x9f, 0x2e, 0xd3, 0x04,
	0xcc, 0xf3, 0x9d, 0x16, 0x43, 0x83, 0x3e, 0x53, 0x26, 0x9f, 0xe5, 0x01, 0x75, 0x61, 0xc5, 0x68,
	0xc9, 0xfc, 0x6c, 0x99, 0xb2, 0x35, 0x8f, 0x18, 0x96, 0xe9, 0xc5, 0x44, 0xba, 0x6c, 0x7c, 0x62,
	0xd2, 0x35, 0xcb, 0x42, 0x89, 0x10, 0x8d, 0x1f, 0xb4, 0x4f, 0x4e, 0xba, 0x86, 0x00, 0xd9, 0xdb,
	0xe9, 0x73, 0x93, 0xae, 0x21, 0x9f, 0xd4, 0x96, 0x9a, 0x15, 0x13, 0xa6, 0xf0, 0xf9, 0x32, 0xad,
	0xd9, 0x79, 0xcc, 0x62, 0xa7, 0x61, 0x8c, 0x36, 0x70, 0x75, 0x0f, 0xbf, 0x53, 0x07, 0x7e, 0xaa,
	0x4c, 0x1b, 0xd0, 0x78, 0xf4, 0xaa, 0x22, 0xd8, 0xe8, 0x27, 0xdb, 0x11, 0x7c, 0x7a, 0x92, 0x35,
	0xcd, 0xfa, 0x39, 0x54, 0x8b, 0x22, 0x8a, 0x51, 0xc2, 0x33, 0x93, 0x74, 0xad, 0x19, 0xf4, 0x89,
	0x44, 0x98, 0x6b, 0x93, 0x2c, 0xce, 0xdc, 0x45, 0x88, 0xeb, 0xe5, 0xea, 0xbb, 0xb7, 0xde, 0x98,
	0xbb, 0xe3, 0xa5, 0xdb, 0x73, 0x85, 0xad, 0xdb, 0x73, 0x85, 0x3f, 0xdd, 0x9e, 0x2b, 0x7c, 0xa0,
	0x94, 0xbd, 0xc6, 0x2d, 0x06, 0xdd, 0x05, 0xff, 0x77, 0x21, 0xd4, 0x0b, 0xc9, 0x46, 0xb8, 0x40,
	0xef, 0xf3, 0x4b, 0xfb, 0xfd, 0x9b, 0xfe, 0xbd, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x6e, 0xe3,
	0x78, 0xbc, 0x09, 0x10, 0x00, 0x00,
}

func (m *ErrDetails) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ErrDetails) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ErrDetails) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Codes) > 0 {
		dAtA2 := make([]byte, len(m.Codes)*10)
		var j1 int
		for _, num := range m.Codes {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintErrcode(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintErrcode(dAtA []byte, offset int, v uint64) int {
	offset -= sovErrcode(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func (m *ErrDetails) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Codes) > 0 {
		l = 0
		for _, e := range m.Codes {
			l += sovErrcode(uint64(e))
		}
		n += 1 + sovErrcode(uint64(l)) + l
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovErrcode(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func sozErrcode(x uint64) (n int) {
	return sovErrcode(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}

func (m *ErrDetails) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowErrcode
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ErrDetails: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ErrDetails: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v ErrCode
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowErrcode
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= ErrCode(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Codes = append(m.Codes, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowErrcode
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthErrcode
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthErrcode
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Codes) == 0 {
					m.Codes = make([]ErrCode, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v ErrCode
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowErrcode
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= ErrCode(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Codes = append(m.Codes, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Codes", wireType)
			}
		default:
			iNdEx = preIndex
			skippy, err := skipErrcode(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthErrcode
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func skipErrcode(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowErrcode
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowErrcode
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowErrcode
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthErrcode
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupErrcode
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthErrcode
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthErrcode        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowErrcode          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupErrcode = fmt.Errorf("proto: unexpected end of group")
)
