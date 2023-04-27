// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: dsa_ds.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DSA2DS_MsgID_MsgID int32

const (
	DSA2DS_MsgID_QueryDsStatus DSA2DS_MsgID_MsgID = 0
	DSA2DS_MsgID_CreateRealm   DSA2DS_MsgID_MsgID = 1 //创建房间
)

// Enum value maps for DSA2DS_MsgID_MsgID.
var (
	DSA2DS_MsgID_MsgID_name = map[int32]string{
		0: "QueryDsStatus",
		1: "CreateRealm",
	}
	DSA2DS_MsgID_MsgID_value = map[string]int32{
		"QueryDsStatus": 0,
		"CreateRealm":   1,
	}
)

func (x DSA2DS_MsgID_MsgID) Enum() *DSA2DS_MsgID_MsgID {
	p := new(DSA2DS_MsgID_MsgID)
	*p = x
	return p
}

func (x DSA2DS_MsgID_MsgID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DSA2DS_MsgID_MsgID) Descriptor() protoreflect.EnumDescriptor {
	return file_dsa_ds_proto_enumTypes[0].Descriptor()
}

func (DSA2DS_MsgID_MsgID) Type() protoreflect.EnumType {
	return &file_dsa_ds_proto_enumTypes[0]
}

func (x DSA2DS_MsgID_MsgID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DSA2DS_MsgID_MsgID.Descriptor instead.
func (DSA2DS_MsgID_MsgID) EnumDescriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{0, 0}
}

type DS2DSA_MsgID_MsgID int32

const (
	DS2DSA_MsgID_DSUpdateState   DS2DSA_MsgID_MsgID = 0 //当前游戏状态
	DS2DSA_MsgID_DSLoadOK        DS2DSA_MsgID_MsgID = 1 //进程启动成功
	DS2DSA_MsgID_DSGameEnd       DS2DSA_MsgID_MsgID = 2 //游戏结束，可以回收进程
	DS2DSA_MsgID_DSRealmCreateOk DS2DSA_MsgID_MsgID = 3 //房间创建完成
)

// Enum value maps for DS2DSA_MsgID_MsgID.
var (
	DS2DSA_MsgID_MsgID_name = map[int32]string{
		0: "DSUpdateState",
		1: "DSLoadOK",
		2: "DSGameEnd",
		3: "DSRealmCreateOk",
	}
	DS2DSA_MsgID_MsgID_value = map[string]int32{
		"DSUpdateState":   0,
		"DSLoadOK":        1,
		"DSGameEnd":       2,
		"DSRealmCreateOk": 3,
	}
)

func (x DS2DSA_MsgID_MsgID) Enum() *DS2DSA_MsgID_MsgID {
	p := new(DS2DSA_MsgID_MsgID)
	*p = x
	return p
}

func (x DS2DSA_MsgID_MsgID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DS2DSA_MsgID_MsgID) Descriptor() protoreflect.EnumDescriptor {
	return file_dsa_ds_proto_enumTypes[1].Descriptor()
}

func (DS2DSA_MsgID_MsgID) Type() protoreflect.EnumType {
	return &file_dsa_ds_proto_enumTypes[1]
}

func (x DS2DSA_MsgID_MsgID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DS2DSA_MsgID_MsgID.Descriptor instead.
func (DS2DSA_MsgID_MsgID) EnumDescriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{3, 0}
}

//
//dsa to ds id
type DSA2DS_MsgID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DSA2DS_MsgID) Reset() {
	*x = DSA2DS_MsgID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DSA2DS_MsgID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DSA2DS_MsgID) ProtoMessage() {}

func (x *DSA2DS_MsgID) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DSA2DS_MsgID.ProtoReflect.Descriptor instead.
func (*DSA2DS_MsgID) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{0}
}

//*
//DSA2DS_MsgIDQueryDsStatus
type QueryDSStatusReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *QueryDSStatusReq) Reset() {
	*x = QueryDSStatusReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryDSStatusReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryDSStatusReq) ProtoMessage() {}

func (x *QueryDSStatusReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryDSStatusReq.ProtoReflect.Descriptor instead.
func (*QueryDSStatusReq) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{1}
}

//*
//DSA2DS_MsgIDCreateRealm
type RealmCreateReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RealmId string `protobuf:"bytes,1,opt,name=realmId,proto3" json:"realmId,omitempty"`
}

func (x *RealmCreateReq) Reset() {
	*x = RealmCreateReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RealmCreateReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RealmCreateReq) ProtoMessage() {}

func (x *RealmCreateReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RealmCreateReq.ProtoReflect.Descriptor instead.
func (*RealmCreateReq) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{2}
}

func (x *RealmCreateReq) GetRealmId() string {
	if x != nil {
		return x.RealmId
	}
	return ""
}

//
//ds to dsa id
type DS2DSA_MsgID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DS2DSA_MsgID) Reset() {
	*x = DS2DSA_MsgID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DS2DSA_MsgID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DS2DSA_MsgID) ProtoMessage() {}

func (x *DS2DSA_MsgID) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DS2DSA_MsgID.ProtoReflect.Descriptor instead.
func (*DS2DSA_MsgID) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{3}
}

//*
//DS2DSA_MsgIDDSLoadOK
type DSLoadOKReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DsId string `protobuf:"bytes,1,opt,name=dsId,proto3" json:"dsId,omitempty"` //DS id
}

func (x *DSLoadOKReq) Reset() {
	*x = DSLoadOKReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DSLoadOKReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DSLoadOKReq) ProtoMessage() {}

func (x *DSLoadOKReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DSLoadOKReq.ProtoReflect.Descriptor instead.
func (*DSLoadOKReq) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{4}
}

func (x *DSLoadOKReq) GetDsId() string {
	if x != nil {
		return x.DsId
	}
	return ""
}

//*
//DS2DSA_MsgIDDSGameEnd
type DSGameEndReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DsId string `protobuf:"bytes,1,opt,name=dsId,proto3" json:"dsId,omitempty"`
}

func (x *DSGameEndReq) Reset() {
	*x = DSGameEndReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DSGameEndReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DSGameEndReq) ProtoMessage() {}

func (x *DSGameEndReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DSGameEndReq.ProtoReflect.Descriptor instead.
func (*DSGameEndReq) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{5}
}

func (x *DSGameEndReq) GetDsId() string {
	if x != nil {
		return x.DsId
	}
	return ""
}

//*
//DS2DSA_MsgIDDSLoadOK
type DSUpdateStateReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DSUpdateStateReq) Reset() {
	*x = DSUpdateStateReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DSUpdateStateReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DSUpdateStateReq) ProtoMessage() {}

func (x *DSUpdateStateReq) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DSUpdateStateReq.ProtoReflect.Descriptor instead.
func (*DSUpdateStateReq) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{6}
}

//*
//DS2DSA_MsgIDDSRealmCreateOk
type RealmCreateResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RealmId string `protobuf:"bytes,1,opt,name=realmId,proto3" json:"realmId,omitempty"`
	DsId    string `protobuf:"bytes,2,opt,name=dsId,proto3" json:"dsId,omitempty"`
}

func (x *RealmCreateResp) Reset() {
	*x = RealmCreateResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsa_ds_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RealmCreateResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RealmCreateResp) ProtoMessage() {}

func (x *RealmCreateResp) ProtoReflect() protoreflect.Message {
	mi := &file_dsa_ds_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RealmCreateResp.ProtoReflect.Descriptor instead.
func (*RealmCreateResp) Descriptor() ([]byte, []int) {
	return file_dsa_ds_proto_rawDescGZIP(), []int{7}
}

func (x *RealmCreateResp) GetRealmId() string {
	if x != nil {
		return x.RealmId
	}
	return ""
}

func (x *RealmCreateResp) GetDsId() string {
	if x != nil {
		return x.DsId
	}
	return ""
}

var File_dsa_ds_proto protoreflect.FileDescriptor

var file_dsa_ds_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x64, 0x73, 0x61, 0x5f, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02,
	0x70, 0x62, 0x22, 0x3b, 0x0a, 0x0c, 0x44, 0x53, 0x41, 0x32, 0x44, 0x53, 0x5f, 0x4d, 0x73, 0x67,
	0x49, 0x44, 0x22, 0x2b, 0x0a, 0x05, 0x4d, 0x73, 0x67, 0x49, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x44, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x10, 0x00, 0x12, 0x0f,
	0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x61, 0x6c, 0x6d, 0x10, 0x01, 0x22,
	0x12, 0x0a, 0x10, 0x51, 0x75, 0x65, 0x72, 0x79, 0x44, 0x53, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x65, 0x71, 0x22, 0x2a, 0x0a, 0x0e, 0x52, 0x65, 0x61, 0x6c, 0x6d, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x49, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x49, 0x64, 0x22,
	0x5c, 0x0a, 0x0c, 0x44, 0x53, 0x32, 0x44, 0x53, 0x41, 0x5f, 0x4d, 0x73, 0x67, 0x49, 0x44, 0x22,
	0x4c, 0x0a, 0x05, 0x4d, 0x73, 0x67, 0x49, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x44, 0x53, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x44,
	0x53, 0x4c, 0x6f, 0x61, 0x64, 0x4f, 0x4b, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x44, 0x53, 0x47,
	0x61, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x10, 0x02, 0x12, 0x13, 0x0a, 0x0f, 0x44, 0x53, 0x52, 0x65,
	0x61, 0x6c, 0x6d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4f, 0x6b, 0x10, 0x03, 0x22, 0x21, 0x0a,
	0x0b, 0x44, 0x53, 0x4c, 0x6f, 0x61, 0x64, 0x4f, 0x4b, 0x52, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x73, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x73, 0x49, 0x64,
	0x22, 0x22, 0x0a, 0x0c, 0x44, 0x53, 0x47, 0x61, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x52, 0x65, 0x71,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x73, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x64, 0x73, 0x49, 0x64, 0x22, 0x12, 0x0a, 0x10, 0x44, 0x53, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x22, 0x3f, 0x0a, 0x0f, 0x52, 0x65, 0x61, 0x6c,
	0x6d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x72,
	0x65, 0x61, 0x6c, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65,
	0x61, 0x6c, 0x6d, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x73, 0x49, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x73, 0x49, 0x64, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x3b, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dsa_ds_proto_rawDescOnce sync.Once
	file_dsa_ds_proto_rawDescData = file_dsa_ds_proto_rawDesc
)

func file_dsa_ds_proto_rawDescGZIP() []byte {
	file_dsa_ds_proto_rawDescOnce.Do(func() {
		file_dsa_ds_proto_rawDescData = protoimpl.X.CompressGZIP(file_dsa_ds_proto_rawDescData)
	})
	return file_dsa_ds_proto_rawDescData
}

var file_dsa_ds_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_dsa_ds_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_dsa_ds_proto_goTypes = []interface{}{
	(DSA2DS_MsgID_MsgID)(0),  // 0: pb.DSA2DS_MsgID.MsgID
	(DS2DSA_MsgID_MsgID)(0),  // 1: pb.DS2DSA_MsgID.MsgID
	(*DSA2DS_MsgID)(nil),     // 2: pb.DSA2DS_MsgID
	(*QueryDSStatusReq)(nil), // 3: pb.QueryDSStatusReq
	(*RealmCreateReq)(nil),   // 4: pb.RealmCreateReq
	(*DS2DSA_MsgID)(nil),     // 5: pb.DS2DSA_MsgID
	(*DSLoadOKReq)(nil),      // 6: pb.DSLoadOKReq
	(*DSGameEndReq)(nil),     // 7: pb.DSGameEndReq
	(*DSUpdateStateReq)(nil), // 8: pb.DSUpdateStateReq
	(*RealmCreateResp)(nil),  // 9: pb.RealmCreateResp
}
var file_dsa_ds_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_dsa_ds_proto_init() }
func file_dsa_ds_proto_init() {
	if File_dsa_ds_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dsa_ds_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DSA2DS_MsgID); i {
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
		file_dsa_ds_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryDSStatusReq); i {
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
		file_dsa_ds_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RealmCreateReq); i {
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
		file_dsa_ds_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DS2DSA_MsgID); i {
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
		file_dsa_ds_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DSLoadOKReq); i {
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
		file_dsa_ds_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DSGameEndReq); i {
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
		file_dsa_ds_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DSUpdateStateReq); i {
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
		file_dsa_ds_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RealmCreateResp); i {
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
			RawDescriptor: file_dsa_ds_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dsa_ds_proto_goTypes,
		DependencyIndexes: file_dsa_ds_proto_depIdxs,
		EnumInfos:         file_dsa_ds_proto_enumTypes,
		MessageInfos:      file_dsa_ds_proto_msgTypes,
	}.Build()
	File_dsa_ds_proto = out.File
	file_dsa_ds_proto_rawDesc = nil
	file_dsa_ds_proto_goTypes = nil
	file_dsa_ds_proto_depIdxs = nil
}