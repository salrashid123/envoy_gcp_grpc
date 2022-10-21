package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const ()

var (
	projectID = flag.String("projectID", "fabled-ray-104117", "projectID")
)

func main() {

	flag.Parse()
	ctx := context.Background()

	pemServerCA, err := ioutil.ReadFile("../certs/tls-ca-chain.pem")
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		panic(err)
	}

	config := &tls.Config{
		RootCAs:    certPool,
		ServerName: "pubsub.googleapis.com",
	}

	tlsCredentials := credentials.NewTLS(config)

	client, err := pubsub.NewClient(ctx, *projectID, option.WithEndpoint("localhost:8081"), option.WithGRPCDialOption(
		grpc.WithTransportCredentials(tlsCredentials),
	))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	t := client.Topic("topic1")

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		result := t.Publish(ctx, &pubsub.Message{
			Data: []byte(fmt.Sprintf("foo number %d", i)),
		})
		wg.Add(1)
		go func(res *pubsub.PublishResult) {
			defer wg.Done()
			id, err := res.Get(ctx)
			if err != nil {
				fmt.Printf("Failed to publish: %v\n", err)
				return
			}
			fmt.Printf("Published message msg ID: %v\n", id)
		}(result)
	}

	wg.Wait()

}
