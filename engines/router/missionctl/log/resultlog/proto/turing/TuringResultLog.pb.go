// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.14.0
// source: TuringResultLog.proto

package turing

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The header from the incoming request to the Turing router. The map value is a comma-delimited string.
	Header map[string]string `protobuf:"bytes,1,rep,name=header,proto3" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The JSON body of the request to the Turing router
	Body string `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_TuringResultLog_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_TuringResultLog_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_TuringResultLog_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetHeader() map[string]string {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Request) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The response from a Turing component (Enricher / Experiment Engine / Router / Ensembler).
	Response string `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	// The error from a Turing component, when a successful response is not received.
	Error string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_TuringResultLog_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_TuringResultLog_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_TuringResultLog_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetResponse() string {
	if x != nil {
		return x.Response
	}
	return ""
}

func (x *Response) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

// key
type TuringResultLogKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique request id generated by Turing, for every incoming request to the Turing router
	TuringReqId string `protobuf:"bytes,1,opt,name=turing_req_id,json=turingReqId,proto3" json:"turing_req_id,omitempty"`
	// The time at which the final response from the Turing router is generated
	EventTimestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=event_timestamp,json=eventTimestamp,proto3" json:"event_timestamp,omitempty"`
}

func (x *TuringResultLogKey) Reset() {
	*x = TuringResultLogKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_TuringResultLog_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TuringResultLogKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TuringResultLogKey) ProtoMessage() {}

func (x *TuringResultLogKey) ProtoReflect() protoreflect.Message {
	mi := &file_TuringResultLog_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TuringResultLogKey.ProtoReflect.Descriptor instead.
func (*TuringResultLogKey) Descriptor() ([]byte, []int) {
	return file_TuringResultLog_proto_rawDescGZIP(), []int{2}
}

func (x *TuringResultLogKey) GetTuringReqId() string {
	if x != nil {
		return x.TuringReqId
	}
	return ""
}

func (x *TuringResultLogKey) GetEventTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.EventTimestamp
	}
	return nil
}

// message
type TuringResultLogMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique request id generated by Turing, for every incoming request to the Turing router
	TuringReqId string `protobuf:"bytes,1,opt,name=turing_req_id,json=turingReqId,proto3" json:"turing_req_id,omitempty"`
	// The time at which the final response from the Turing router is generated
	EventTimestamp *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=event_timestamp,json=eventTimestamp,proto3" json:"event_timestamp,omitempty"`
	// The version of the router deployed. This corresponds to the name and version of the router deployed from the Turing app.
	// Format: {router_name}-{router_version}.{project_name}
	RouterVersion string `protobuf:"bytes,3,opt,name=router_version,json=routerVersion,proto3" json:"router_version,omitempty"`
	// The original request to the Turing router
	Request *Request `protobuf:"bytes,4,opt,name=request,proto3" json:"request,omitempty"`
	// The response from the Experiment engine, if configured
	Experiment *Response `protobuf:"bytes,5,opt,name=experiment,proto3" json:"experiment,omitempty"`
	// The response from the Enricher, if configured
	Enricher *Response `protobuf:"bytes,6,opt,name=enricher,proto3" json:"enricher,omitempty"`
	// The response from the routes
	Router *Response `protobuf:"bytes,7,opt,name=router,proto3" json:"router,omitempty"`
	// The response from the Enricher, if configured
	Ensembler *Response `protobuf:"bytes,8,opt,name=ensembler,proto3" json:"ensembler,omitempty"`
}

func (x *TuringResultLogMessage) Reset() {
	*x = TuringResultLogMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_TuringResultLog_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TuringResultLogMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TuringResultLogMessage) ProtoMessage() {}

func (x *TuringResultLogMessage) ProtoReflect() protoreflect.Message {
	mi := &file_TuringResultLog_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TuringResultLogMessage.ProtoReflect.Descriptor instead.
func (*TuringResultLogMessage) Descriptor() ([]byte, []int) {
	return file_TuringResultLog_proto_rawDescGZIP(), []int{3}
}

func (x *TuringResultLogMessage) GetTuringReqId() string {
	if x != nil {
		return x.TuringReqId
	}
	return ""
}

func (x *TuringResultLogMessage) GetEventTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.EventTimestamp
	}
	return nil
}

func (x *TuringResultLogMessage) GetRouterVersion() string {
	if x != nil {
		return x.RouterVersion
	}
	return ""
}

func (x *TuringResultLogMessage) GetRequest() *Request {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *TuringResultLogMessage) GetExperiment() *Response {
	if x != nil {
		return x.Experiment
	}
	return nil
}

func (x *TuringResultLogMessage) GetEnricher() *Response {
	if x != nil {
		return x.Enricher
	}
	return nil
}

func (x *TuringResultLogMessage) GetRouter() *Response {
	if x != nil {
		return x.Router
	}
	return nil
}

func (x *TuringResultLogMessage) GetEnsembler() *Response {
	if x != nil {
		return x.Ensembler
	}
	return nil
}

var File_TuringResultLog_proto protoreflect.FileDescriptor

var file_TuringResultLog_proto_rawDesc = []byte{
	0x0a, 0x15, 0x54, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x6f,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x8d, 0x01, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x06,
	0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74,
	0x75, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x62, 0x6f, 0x64, 0x79, 0x1a, 0x39, 0x0a, 0x0b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0x3c, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x7d,
	0x0a, 0x12, 0x54, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x6f,
	0x67, 0x4b, 0x65, 0x79, 0x12, 0x22, 0x0a, 0x0d, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x72,
	0x65, 0x71, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x74, 0x75, 0x72,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x49, 0x64, 0x12, 0x43, 0x0a, 0x0f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x8d, 0x03,
	0x0a, 0x16, 0x54, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x6f,
	0x67, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x22, 0x0a, 0x0d, 0x74, 0x75, 0x72, 0x69,
	0x6e, 0x67, 0x5f, 0x72, 0x65, 0x71, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x49, 0x64, 0x12, 0x43, 0x0a, 0x0f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x74, 0x75, 0x72, 0x69,
	0x6e, 0x67, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65, 0x6e,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x65, 0x72,
	0x69, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x2c, 0x0a, 0x08, 0x65, 0x6e, 0x72, 0x69, 0x63, 0x68, 0x65,
	0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x65, 0x6e, 0x72, 0x69, 0x63,
	0x68, 0x65, 0x72, 0x12, 0x28, 0x0a, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x12, 0x2e, 0x0a,
	0x09, 0x65, 0x6e, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x52, 0x09, 0x65, 0x6e, 0x73, 0x65, 0x6d, 0x62, 0x6c, 0x65, 0x72, 0x42, 0x2f, 0x42,
	0x14, 0x54, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x6f, 0x67,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x5a, 0x17, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x6f, 0x6a, 0x65, 0x6b, 0x2f, 0x74, 0x75, 0x72, 0x69, 0x6e, 0x67, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_TuringResultLog_proto_rawDescOnce sync.Once
	file_TuringResultLog_proto_rawDescData = file_TuringResultLog_proto_rawDesc
)

func file_TuringResultLog_proto_rawDescGZIP() []byte {
	file_TuringResultLog_proto_rawDescOnce.Do(func() {
		file_TuringResultLog_proto_rawDescData = protoimpl.X.CompressGZIP(file_TuringResultLog_proto_rawDescData)
	})
	return file_TuringResultLog_proto_rawDescData
}

var file_TuringResultLog_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_TuringResultLog_proto_goTypes = []interface{}{
	(*Request)(nil),                // 0: turing.Request
	(*Response)(nil),               // 1: turing.Response
	(*TuringResultLogKey)(nil),     // 2: turing.TuringResultLogKey
	(*TuringResultLogMessage)(nil), // 3: turing.TuringResultLogMessage
	nil,                            // 4: turing.Request.HeaderEntry
	(*timestamppb.Timestamp)(nil),  // 5: google.protobuf.Timestamp
}
var file_TuringResultLog_proto_depIdxs = []int32{
	4, // 0: turing.Request.header:type_name -> turing.Request.HeaderEntry
	5, // 1: turing.TuringResultLogKey.event_timestamp:type_name -> google.protobuf.Timestamp
	5, // 2: turing.TuringResultLogMessage.event_timestamp:type_name -> google.protobuf.Timestamp
	0, // 3: turing.TuringResultLogMessage.request:type_name -> turing.Request
	1, // 4: turing.TuringResultLogMessage.experiment:type_name -> turing.Response
	1, // 5: turing.TuringResultLogMessage.enricher:type_name -> turing.Response
	1, // 6: turing.TuringResultLogMessage.router:type_name -> turing.Response
	1, // 7: turing.TuringResultLogMessage.ensembler:type_name -> turing.Response
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_TuringResultLog_proto_init() }
func file_TuringResultLog_proto_init() {
	if File_TuringResultLog_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_TuringResultLog_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_TuringResultLog_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_TuringResultLog_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TuringResultLogKey); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_TuringResultLog_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TuringResultLogMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_TuringResultLog_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_TuringResultLog_proto_goTypes,
		DependencyIndexes: file_TuringResultLog_proto_depIdxs,
		MessageInfos:      file_TuringResultLog_proto_msgTypes,
	}.Build()
	File_TuringResultLog_proto = out.File
	file_TuringResultLog_proto_rawDesc = nil
	file_TuringResultLog_proto_goTypes = nil
	file_TuringResultLog_proto_depIdxs = nil
}
