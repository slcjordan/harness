package main

import (
	"context"
	"fmt"

	"github.com/fullstorydev/grpcui"
	"github.com/fullstorydev/grpcui/standalone"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/slcjordan/harness/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getMethods(source grpcurl.DescriptorSource) ([]*desc.MethodDescriptor, error) {
	allServices, err := source.ListServices()
	if err != nil {
		return nil, err
	}

	var descs []*desc.MethodDescriptor
	for _, svc := range allServices {
		if svc == "grpc.reflection.v1alpha.ServerReflection" || svc == "grpc.reflection.v1.ServerReflection" {
			continue
		}
		d, err := source.FindSymbol(svc)
		if err != nil {
			return nil, err
		}
		sd, ok := d.(*desc.ServiceDescriptor)
		if !ok {
			return nil, fmt.Errorf("%s should be a service descriptor but instead is a %T", d.GetFullyQualifiedName(), d)
		}
		for _, md := range sd.GetMethods() {
			descs = append(descs, md)
		}
	}

	return descs, nil
}

func main() {
	logger.Init()
	src, err := grpcurl.DescriptorSourceFromProtoFiles([]string{"/home/jcrabtree/Developer/src/gitlab.com/vivint/horizontals/platform/device"})
	if err != nil {
		fmt.Println("DescriptorSourceFromProtoFiles", err)
		return
	}
	methods, err := getMethods(src)
	if err != nil {
		fmt.Println("getMethods", err)
		return
	}
	cc, err := grpc.DialContext(context.TODO(), "dns:///device:8856", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("dialcontext", err)
		return
	}
	f, err := grpcui.AllFilesViaReflection(context.TODO(), cc)
	if err != nil {
		fmt.Println("AllFilesViaReflection", err)
		return
	}
	standalone.Handler(cc, "dns:///device:8856", methods, f)
	/*
		if err != nil {
			fmt.Println("handlerviareflection", err)
			return
		}
	*/
}
