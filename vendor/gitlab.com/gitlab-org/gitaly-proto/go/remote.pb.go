// Code generated by protoc-gen-go. DO NOT EDIT.
// source: remote.proto

package gitaly

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type AddRemoteRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	Name       string      `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Url        string      `protobuf:"bytes,3,opt,name=url" json:"url,omitempty"`
	// If set the remote is configured as a mirror with that mapping
	MirrorRefmap string `protobuf:"bytes,4,opt,name=mirror_refmap,json=mirrorRefmap" json:"mirror_refmap,omitempty"`
}

func (m *AddRemoteRequest) Reset()                    { *m = AddRemoteRequest{} }
func (m *AddRemoteRequest) String() string            { return proto.CompactTextString(m) }
func (*AddRemoteRequest) ProtoMessage()               {}
func (*AddRemoteRequest) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{0} }

func (m *AddRemoteRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *AddRemoteRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *AddRemoteRequest) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *AddRemoteRequest) GetMirrorRefmap() string {
	if m != nil {
		return m.MirrorRefmap
	}
	return ""
}

type AddRemoteResponse struct {
}

func (m *AddRemoteResponse) Reset()                    { *m = AddRemoteResponse{} }
func (m *AddRemoteResponse) String() string            { return proto.CompactTextString(m) }
func (*AddRemoteResponse) ProtoMessage()               {}
func (*AddRemoteResponse) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{1} }

type RemoveRemoteRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	Name       string      `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
}

func (m *RemoveRemoteRequest) Reset()                    { *m = RemoveRemoteRequest{} }
func (m *RemoveRemoteRequest) String() string            { return proto.CompactTextString(m) }
func (*RemoveRemoteRequest) ProtoMessage()               {}
func (*RemoveRemoteRequest) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{2} }

func (m *RemoveRemoteRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *RemoveRemoteRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type RemoveRemoteResponse struct {
	Result bool `protobuf:"varint,1,opt,name=result" json:"result,omitempty"`
}

func (m *RemoveRemoteResponse) Reset()                    { *m = RemoveRemoteResponse{} }
func (m *RemoveRemoteResponse) String() string            { return proto.CompactTextString(m) }
func (*RemoveRemoteResponse) ProtoMessage()               {}
func (*RemoveRemoteResponse) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{3} }

func (m *RemoveRemoteResponse) GetResult() bool {
	if m != nil {
		return m.Result
	}
	return false
}

type FetchInternalRemoteRequest struct {
	Repository       *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	RemoteRepository *Repository `protobuf:"bytes,2,opt,name=remote_repository,json=remoteRepository" json:"remote_repository,omitempty"`
}

func (m *FetchInternalRemoteRequest) Reset()                    { *m = FetchInternalRemoteRequest{} }
func (m *FetchInternalRemoteRequest) String() string            { return proto.CompactTextString(m) }
func (*FetchInternalRemoteRequest) ProtoMessage()               {}
func (*FetchInternalRemoteRequest) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{4} }

func (m *FetchInternalRemoteRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *FetchInternalRemoteRequest) GetRemoteRepository() *Repository {
	if m != nil {
		return m.RemoteRepository
	}
	return nil
}

type FetchInternalRemoteResponse struct {
	Result bool `protobuf:"varint,1,opt,name=result" json:"result,omitempty"`
}

func (m *FetchInternalRemoteResponse) Reset()                    { *m = FetchInternalRemoteResponse{} }
func (m *FetchInternalRemoteResponse) String() string            { return proto.CompactTextString(m) }
func (*FetchInternalRemoteResponse) ProtoMessage()               {}
func (*FetchInternalRemoteResponse) Descriptor() ([]byte, []int) { return fileDescriptor9, []int{5} }

func (m *FetchInternalRemoteResponse) GetResult() bool {
	if m != nil {
		return m.Result
	}
	return false
}

func init() {
	proto.RegisterType((*AddRemoteRequest)(nil), "gitaly.AddRemoteRequest")
	proto.RegisterType((*AddRemoteResponse)(nil), "gitaly.AddRemoteResponse")
	proto.RegisterType((*RemoveRemoteRequest)(nil), "gitaly.RemoveRemoteRequest")
	proto.RegisterType((*RemoveRemoteResponse)(nil), "gitaly.RemoveRemoteResponse")
	proto.RegisterType((*FetchInternalRemoteRequest)(nil), "gitaly.FetchInternalRemoteRequest")
	proto.RegisterType((*FetchInternalRemoteResponse)(nil), "gitaly.FetchInternalRemoteResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for RemoteService service

type RemoteServiceClient interface {
	AddRemote(ctx context.Context, in *AddRemoteRequest, opts ...grpc.CallOption) (*AddRemoteResponse, error)
	FetchInternalRemote(ctx context.Context, in *FetchInternalRemoteRequest, opts ...grpc.CallOption) (*FetchInternalRemoteResponse, error)
	RemoveRemote(ctx context.Context, in *RemoveRemoteRequest, opts ...grpc.CallOption) (*RemoveRemoteResponse, error)
}

type remoteServiceClient struct {
	cc *grpc.ClientConn
}

func NewRemoteServiceClient(cc *grpc.ClientConn) RemoteServiceClient {
	return &remoteServiceClient{cc}
}

func (c *remoteServiceClient) AddRemote(ctx context.Context, in *AddRemoteRequest, opts ...grpc.CallOption) (*AddRemoteResponse, error) {
	out := new(AddRemoteResponse)
	err := grpc.Invoke(ctx, "/gitaly.RemoteService/AddRemote", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteServiceClient) FetchInternalRemote(ctx context.Context, in *FetchInternalRemoteRequest, opts ...grpc.CallOption) (*FetchInternalRemoteResponse, error) {
	out := new(FetchInternalRemoteResponse)
	err := grpc.Invoke(ctx, "/gitaly.RemoteService/FetchInternalRemote", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteServiceClient) RemoveRemote(ctx context.Context, in *RemoveRemoteRequest, opts ...grpc.CallOption) (*RemoveRemoteResponse, error) {
	out := new(RemoveRemoteResponse)
	err := grpc.Invoke(ctx, "/gitaly.RemoteService/RemoveRemote", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RemoteService service

type RemoteServiceServer interface {
	AddRemote(context.Context, *AddRemoteRequest) (*AddRemoteResponse, error)
	FetchInternalRemote(context.Context, *FetchInternalRemoteRequest) (*FetchInternalRemoteResponse, error)
	RemoveRemote(context.Context, *RemoveRemoteRequest) (*RemoveRemoteResponse, error)
}

func RegisterRemoteServiceServer(s *grpc.Server, srv RemoteServiceServer) {
	s.RegisterService(&_RemoteService_serviceDesc, srv)
}

func _RemoteService_AddRemote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddRemoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteServiceServer).AddRemote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.RemoteService/AddRemote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteServiceServer).AddRemote(ctx, req.(*AddRemoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteService_FetchInternalRemote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchInternalRemoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteServiceServer).FetchInternalRemote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.RemoteService/FetchInternalRemote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteServiceServer).FetchInternalRemote(ctx, req.(*FetchInternalRemoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RemoteService_RemoveRemote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveRemoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemoteServiceServer).RemoveRemote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.RemoteService/RemoveRemote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemoteServiceServer).RemoveRemote(ctx, req.(*RemoveRemoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RemoteService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.RemoteService",
	HandlerType: (*RemoteServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddRemote",
			Handler:    _RemoteService_AddRemote_Handler,
		},
		{
			MethodName: "FetchInternalRemote",
			Handler:    _RemoteService_FetchInternalRemote_Handler,
		},
		{
			MethodName: "RemoveRemote",
			Handler:    _RemoteService_RemoveRemote_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "remote.proto",
}

func init() { proto.RegisterFile("remote.proto", fileDescriptor9) }

var fileDescriptor9 = []byte{
	// 323 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0x41, 0x4f, 0xf2, 0x40,
	0x10, 0xfd, 0x0a, 0x84, 0x7c, 0x8c, 0x25, 0x81, 0xc1, 0x98, 0x5a, 0x3c, 0x90, 0x72, 0xe1, 0xd4,
	0x03, 0xc6, 0xb3, 0xd1, 0x83, 0x89, 0xf1, 0xb6, 0x9e, 0x0d, 0x56, 0x18, 0xa5, 0x49, 0xdb, 0xad,
	0xb3, 0x0b, 0x09, 0x57, 0xff, 0x81, 0xff, 0xd8, 0xb0, 0x4b, 0x6b, 0xd5, 0xc2, 0xc5, 0x78, 0xdb,
	0xbe, 0x99, 0x37, 0xef, 0x75, 0xde, 0x2e, 0xb8, 0x4c, 0xa9, 0xd4, 0x14, 0xe6, 0x2c, 0xb5, 0xc4,
	0xf6, 0x4b, 0xac, 0xa3, 0x64, 0xe3, 0xbb, 0x6a, 0x19, 0x31, 0x2d, 0x2c, 0x1a, 0xbc, 0x3b, 0xd0,
	0xbb, 0x5a, 0x2c, 0x84, 0xe9, 0x14, 0xf4, 0xba, 0x22, 0xa5, 0x71, 0x0a, 0xc0, 0x94, 0x4b, 0x15,
	0x6b, 0xc9, 0x1b, 0xcf, 0x19, 0x39, 0x93, 0xa3, 0x29, 0x86, 0x96, 0x1f, 0x8a, 0xb2, 0x22, 0x2a,
	0x5d, 0x88, 0xd0, 0xca, 0xa2, 0x94, 0xbc, 0xc6, 0xc8, 0x99, 0x74, 0x84, 0x39, 0x63, 0x0f, 0x9a,
	0x2b, 0x4e, 0xbc, 0xa6, 0x81, 0xb6, 0x47, 0x1c, 0x43, 0x37, 0x8d, 0x99, 0x25, 0xcf, 0x98, 0x9e,
	0xd3, 0x28, 0xf7, 0x5a, 0xa6, 0xe6, 0x5a, 0x50, 0x18, 0x2c, 0x18, 0x40, 0xbf, 0x62, 0x49, 0xe5,
	0x32, 0x53, 0x14, 0x3c, 0xc0, 0x60, 0x8b, 0xac, 0xe9, 0x4f, 0xac, 0x06, 0x21, 0x1c, 0x7f, 0x1d,
	0x6f, 0x65, 0xf1, 0x04, 0xda, 0x4c, 0x6a, 0x95, 0x68, 0x33, 0xfb, 0xbf, 0xd8, 0x7d, 0x6d, 0xf7,
	0xe6, 0xdf, 0x90, 0x9e, 0x2f, 0x6f, 0x33, 0x4d, 0x9c, 0x45, 0xc9, 0xef, 0x6d, 0x5d, 0x42, 0xdf,
	0x06, 0x36, 0xab, 0x50, 0x1b, 0x7b, 0xa9, 0x3d, 0xde, 0x29, 0x16, 0x48, 0x70, 0x01, 0xc3, 0x5a,
	0x4b, 0x87, 0x7f, 0x65, 0xfa, 0xd6, 0x80, 0xae, 0x6d, 0xbd, 0x27, 0x5e, 0xc7, 0x73, 0xc2, 0x6b,
	0xe8, 0x94, 0x01, 0xa0, 0x57, 0x68, 0x7f, 0xbf, 0x26, 0xfe, 0x69, 0x4d, 0x65, 0x97, 0xd6, 0x3f,
	0x7c, 0x84, 0x41, 0x8d, 0x19, 0x0c, 0x0a, 0xce, 0xfe, 0xe5, 0xf9, 0xe3, 0x83, 0x3d, 0xa5, 0xc2,
	0x1d, 0xb8, 0xd5, 0xc8, 0x70, 0xf8, 0xb9, 0xa4, 0x1f, 0xf7, 0xc4, 0x3f, 0xab, 0x2f, 0x16, 0xc3,
	0x9e, 0xda, 0xe6, 0x39, 0x9c, 0x7f, 0x04, 0x00, 0x00, 0xff, 0xff, 0x24, 0x14, 0x1d, 0x67, 0x34,
	0x03, 0x00, 0x00,
}