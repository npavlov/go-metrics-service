// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: proto/v1/metrics.proto

package metrics

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MetricService_SetMetrics_FullMethodName = "/proto.metrics.v1.MetricService/SetMetrics"
	MetricService_SetMetric_FullMethodName  = "/proto.metrics.v1.MetricService/SetMetric"
)

// MetricServiceClient is the client API for MetricService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricServiceClient interface {
	SetMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error)
	SetMetric(ctx context.Context, in *MetricRequest, opts ...grpc.CallOption) (*MetricResponse, error)
}

type metricServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricServiceClient(cc grpc.ClientConnInterface) MetricServiceClient {
	return &metricServiceClient{cc}
}

func (c *metricServiceClient) SetMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MetricsResponse)
	err := c.cc.Invoke(ctx, MetricService_SetMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricServiceClient) SetMetric(ctx context.Context, in *MetricRequest, opts ...grpc.CallOption) (*MetricResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MetricResponse)
	err := c.cc.Invoke(ctx, MetricService_SetMetric_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricServiceServer is the server API for MetricService service.
// All implementations must embed UnimplementedMetricServiceServer
// for forward compatibility.
type MetricServiceServer interface {
	SetMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
	SetMetric(context.Context, *MetricRequest) (*MetricResponse, error)
	mustEmbedUnimplementedMetricServiceServer()
}

// UnimplementedMetricServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMetricServiceServer struct{}

func (UnimplementedMetricServiceServer) SetMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMetrics not implemented")
}
func (UnimplementedMetricServiceServer) SetMetric(context.Context, *MetricRequest) (*MetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMetric not implemented")
}
func (UnimplementedMetricServiceServer) mustEmbedUnimplementedMetricServiceServer() {}
func (UnimplementedMetricServiceServer) testEmbeddedByValue()                       {}

// UnsafeMetricServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricServiceServer will
// result in compilation errors.
type UnsafeMetricServiceServer interface {
	mustEmbedUnimplementedMetricServiceServer()
}

func RegisterMetricServiceServer(s grpc.ServiceRegistrar, srv MetricServiceServer) {
	// If the following call pancis, it indicates UnimplementedMetricServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MetricService_ServiceDesc, srv)
}

func _MetricService_SetMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServiceServer).SetMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricService_SetMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServiceServer).SetMetrics(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricService_SetMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServiceServer).SetMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricService_SetMetric_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServiceServer).SetMetric(ctx, req.(*MetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricService_ServiceDesc is the grpc.ServiceDesc for MetricService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetricService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.metrics.v1.MetricService",
	HandlerType: (*MetricServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetMetrics",
			Handler:    _MetricService_SetMetrics_Handler,
		},
		{
			MethodName: "SetMetric",
			Handler:    _MetricService_SetMetric_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/v1/metrics.proto",
}
