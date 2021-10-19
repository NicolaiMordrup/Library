package library

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// RunGRPCGateway initlaizes and starts a grpc gateway
func RunGRPCGateway(
	ctx context.Context,
	log *zap.SugaredLogger,
	addr string,
	grpcAddr string,
	registerHandlersFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error,
) error {
	gatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
		}),
	)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Connect to the gRPC server
	log.Infow(
		"connecting gRPC gateway to the gRPC service",
		"addr", addr,
	)
	dialOption := grpc.WithInsecure()
	conn, err := grpc.DialContext(ctx, addr, dialOption, grpc.WithBlock())
	if err != nil {
		return err
	}

	// Register handlers towards the gRPC server
	if err := registerHandlersFunc(ctx, gatewayMux, conn); err != nil {
		return fmt.Errorf("register handler err, %w", err)
	}

	// Start gateway
	var gatewayServerMux http.Handler = gatewayMux

	server := &http.Server{
		Addr:     grpcAddr,
		Handler:  gatewayServerMux,
		ErrorLog: zap.NewStdLog(log.Desugar()),
	}

	// Run the server
	log.Infow("running REST gateway", "address", grpcAddr)
	errch := make(chan error)
	select {
	case errch <- server.ListenAndServe():
		return <-errch
	case <-ctx.Done():
		if err := server.Close(); err != nil {
			log.Info(err)
		}
		return ctx.Err()
	}
}
