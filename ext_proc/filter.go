package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/psanford/lencode"

	"google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/protobuf/proto"
)

var (
	grpcport = flag.String("grpcport", ":18080", "grpcport")
	tlsCert  = flag.String("tlsCert", "../certs/ext_server.crt", "tls Certificate")
	tlsKey   = flag.String("tlsKey", "../certs/ext_server.key", "tls Key")
)

const ()

type server struct{}

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request + %s", in.String())
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (s *server) Process(srv pb.ExternalProcessor_ProcessServer) error {

	log.Println("Got stream:  -->  ")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		resp := &pb.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *pb.ProcessingRequest_RequestHeaders:
			log.Printf("pb.ProcessingRequest_RequestHeaders %v \n", v)
			r := req.Request
			h := r.(*pb.ProcessingRequest_RequestHeaders)
			log.Printf("Got RequestHeaders.Attributes %v", h.RequestHeaders.Attributes)
			log.Printf("Got RequestHeaders.Headers %v", h.RequestHeaders.Headers)

			for _, n := range h.RequestHeaders.Headers.Headers {
				if n.Key == ":path" && n.Value == "/google.pubsub.v1.Publisher/Publish" {
					for _, n := range h.RequestHeaders.Headers.Headers {
						log.Printf("Header %s %s", n.Key, n.Value)

						rhq := &pb.HeadersResponse{
							Response: &pb.CommonResponse{},
						}

						resp = &pb.ProcessingResponse{
							Response: &pb.ProcessingResponse_RequestHeaders{
								RequestHeaders: rhq,
							},
							ModeOverride: &v3.ProcessingMode{
								RequestBodyMode:    v3.ProcessingMode_BUFFERED,
								ResponseHeaderMode: v3.ProcessingMode_SKIP,
								ResponseBodyMode:   v3.ProcessingMode_NONE,
							},
						}
						break

					}

				}
			}

		case *pb.ProcessingRequest_RequestBody:

			r := req.Request
			b := r.(*pb.ProcessingRequest_RequestBody)
			log.Printf("   RequestBody: %s", string(b.RequestBody.Body))

			if b.RequestBody.EndOfStream {
				dec := lencode.NewDecoder(bytes.NewBuffer(b.RequestBody.Body), lencode.SeparatorOpt([]byte{0}))
				for {
					reqMessageBytes, err := dec.Decode()

					if err != nil {
						if err == io.EOF {
							break
						}
						log.Fatalf("could not Decode  %v", err)
						return err
					}

					serialized := &pubsub.PublishRequest{}
					err = proto.Unmarshal(reqMessageBytes, serialized)
					if err != nil {
						log.Fatal("unmarshaling error: ", err)
					}
					log.Printf(">>>>>>>>>>>>>>>> Got message for topic: %s\n", serialized.Topic)

					for _, m := range serialized.Messages {
						fmt.Printf("Decode PubsubMessage Data ---->  %s\n", string(m.Data))
					}
				}
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_RequestBody{},
					ModeOverride: &v3.ProcessingMode{
						ResponseHeaderMode: v3.ProcessingMode_SEND,
						ResponseBodyMode:   v3.ProcessingMode_NONE,
					},
				}

			}

		case *pb.ProcessingRequest_ResponseHeaders:
			log.Printf("pb.ProcessingRequest_ResponseHeaders %v \n", v)
			rhq := &pb.HeadersResponse{}
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: rhq,
				},
				ModeOverride: &v3.ProcessingMode{
					ResponseBodyMode: v3.ProcessingMode_BUFFERED,
				},
			}

		case *pb.ProcessingRequest_ResponseBody:
			log.Printf("pb.ProcessingRequest_ResponseBody %v \n", v)

			r := req.Request
			b := r.(*pb.ProcessingRequest_ResponseBody)
			log.Printf("   ResponseBody: %s", string(b.ResponseBody.Body))
			resp = &pb.ProcessingResponse{}

		default:
			log.Printf("Unknown Request type %v\n", v)
		}
		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func main() {

	flag.Parse()

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{}

	if *tlsCert == "" || *tlsKey == "" {
		log.Fatalf("Must set --tlsCert and tlsKey")
	}
	ce, err := credentials.NewServerTLSFromFile(*tlsCert, *tlsKey)
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}
	sopts = append(sopts, grpc.Creds(ce))

	s := grpc.NewServer(sopts...)
	pb.RegisterExternalProcessorServer(s, &server{})
	healthpb.RegisterHealthServer(s, &healthServer{})

	log.Printf("Starting server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
