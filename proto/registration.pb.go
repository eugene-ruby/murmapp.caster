// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: proto/registration.proto

package casterpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RegisterWebhookRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	BotId         string                 `protobuf:"bytes,1,opt,name=bot_id,json=botId,proto3" json:"bot_id,omitempty"`
	ApiKeyBot     []byte                 `protobuf:"bytes,2,opt,name=api_key_bot,json=apiKeyBot,proto3" json:"api_key_bot,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RegisterWebhookRequest) Reset() {
	*x = RegisterWebhookRequest{}
	mi := &file_proto_registration_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterWebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterWebhookRequest) ProtoMessage() {}

func (x *RegisterWebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registration_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterWebhookRequest.ProtoReflect.Descriptor instead.
func (*RegisterWebhookRequest) Descriptor() ([]byte, []int) {
	return file_proto_registration_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterWebhookRequest) GetBotId() string {
	if x != nil {
		return x.BotId
	}
	return ""
}

func (x *RegisterWebhookRequest) GetApiKeyBot() []byte {
	if x != nil {
		return x.ApiKeyBot
	}
	return nil
}

type RegisterWebhookResponse struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	BotId              string                 `protobuf:"bytes,1,opt,name=bot_id,json=botId,proto3" json:"bot_id,omitempty"`
	EncryptedApiKeyBot []byte                 `protobuf:"bytes,2,opt,name=encrypted_api_key_bot,json=encryptedApiKeyBot,proto3" json:"encrypted_api_key_bot,omitempty"`
	WebhookId          string                 `protobuf:"bytes,3,opt,name=webhook_id,json=webhookId,proto3" json:"webhook_id,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *RegisterWebhookResponse) Reset() {
	*x = RegisterWebhookResponse{}
	mi := &file_proto_registration_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterWebhookResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterWebhookResponse) ProtoMessage() {}

func (x *RegisterWebhookResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_registration_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterWebhookResponse.ProtoReflect.Descriptor instead.
func (*RegisterWebhookResponse) Descriptor() ([]byte, []int) {
	return file_proto_registration_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterWebhookResponse) GetBotId() string {
	if x != nil {
		return x.BotId
	}
	return ""
}

func (x *RegisterWebhookResponse) GetEncryptedApiKeyBot() []byte {
	if x != nil {
		return x.EncryptedApiKeyBot
	}
	return nil
}

func (x *RegisterWebhookResponse) GetWebhookId() string {
	if x != nil {
		return x.WebhookId
	}
	return ""
}

var File_proto_registration_proto protoreflect.FileDescriptor

const file_proto_registration_proto_rawDesc = "" +
	"\n" +
	"\x18proto/registration.proto\x12\x06caster\"O\n" +
	"\x16RegisterWebhookRequest\x12\x15\n" +
	"\x06bot_id\x18\x01 \x01(\tR\x05botId\x12\x1e\n" +
	"\vapi_key_bot\x18\x02 \x01(\fR\tapiKeyBot\"\x82\x01\n" +
	"\x17RegisterWebhookResponse\x12\x15\n" +
	"\x06bot_id\x18\x01 \x01(\tR\x05botId\x121\n" +
	"\x15encrypted_api_key_bot\x18\x02 \x01(\fR\x12encryptedApiKeyBot\x12\x1d\n" +
	"\n" +
	"webhook_id\x18\x03 \x01(\tR\twebhookIdB\x1fZ\x1dmurmapp.caster/proto;casterpbb\x06proto3"

var (
	file_proto_registration_proto_rawDescOnce sync.Once
	file_proto_registration_proto_rawDescData []byte
)

func file_proto_registration_proto_rawDescGZIP() []byte {
	file_proto_registration_proto_rawDescOnce.Do(func() {
		file_proto_registration_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_registration_proto_rawDesc), len(file_proto_registration_proto_rawDesc)))
	})
	return file_proto_registration_proto_rawDescData
}

var file_proto_registration_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_registration_proto_goTypes = []any{
	(*RegisterWebhookRequest)(nil),  // 0: caster.RegisterWebhookRequest
	(*RegisterWebhookResponse)(nil), // 1: caster.RegisterWebhookResponse
}
var file_proto_registration_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_registration_proto_init() }
func file_proto_registration_proto_init() {
	if File_proto_registration_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_registration_proto_rawDesc), len(file_proto_registration_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_registration_proto_goTypes,
		DependencyIndexes: file_proto_registration_proto_depIdxs,
		MessageInfos:      file_proto_registration_proto_msgTypes,
	}.Build()
	File_proto_registration_proto = out.File
	file_proto_registration_proto_goTypes = nil
	file_proto_registration_proto_depIdxs = nil
}
