package main

import (
	"context"
	"fmt"
	"github.com/aemengo/bosh-runc-cpi/pb"
	"github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"os"
)


var opts struct {
	Target string `short:"t" long:"target" description:"Target the tcp address of a runc cpid server in the following format" value-name:"127.0.0.1:9999"`
}

func main()  {
	args, err := flags.Parse(&opts)
	expectNoError(err)

	conn, err := grpc.Dial(opts.Target, grpc.WithInsecure())
	expectNoError(err)
	defer conn.Close()

	ctx := context.Background()
	cpidClient := pb.NewCPIDClient(conn)

}

func expectNoError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}