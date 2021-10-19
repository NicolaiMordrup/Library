package main

import (
	"context"
	"fmt"
	"os"
	"time"

	library "github.com/NicolaiMordrup/library"
	"github.com/NicolaiMordrup/library/gen/proto/go/librarypb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

func main() {

	// Configuration
	connstr := "./librarystorage.db"
	if envVal := os.Getenv("SQLITE_DB_CONN"); envVal != "" {
		connstr = envVal
	}
	portStr := "8000"
	if envVal := os.Getenv("SERVER_PORT"); envVal != "" {
		portStr = envVal
	}
	minDurationBetweenUpdatesStr := "10s"
	if envVal := os.Getenv("MIN_DURATION_BETWEEN_UPDATES"); envVal != "" {
		minDurationBetweenUpdatesStr = envVal
	}
	minDurationBetweenUpdates, err := time.ParseDuration(minDurationBetweenUpdatesStr)
	check(err, "failed to parse min duration between updates")

	// Setup logger
	structuredLogger, _ := zap.NewProduction()
	log := structuredLogger.Sugar()

	// Connect to database
	db, err := library.NewDB(connstr)
	check(err, "failed to open sqlite connection")
	check(library.EnsureSchema(db), "migration failed")

	// Creating  errgroup such that we can use go routines to start the
	// grpc server and grpc gateway
	g, ctx := errgroup.WithContext(context.Background())
	grpcAddr := ":8001"
	addr := fmt.Sprintf(":%v", portStr)

	// Initialize and starting the grpc Server
	g.Go(func() error {

		myServer := library.NewServer(db, log, minDurationBetweenUpdates)
		log.Infow("starting grpc server",
			"addr", addr,
		)
		return myServer.RunGRPCServer(addr)
	})

	// Initialize and starting the grpc gateway
	g.Go(func() error {
		return library.RunGRPCGateway(ctx, log, addr, grpcAddr,
			func(ctx context.Context, gwMux *runtime.ServeMux, conn *grpc.ClientConn) (retErr error) {
				setErr := func(err error) {
					if retErr == nil {
						retErr = err
					}
				}
				setErr(librarypb.RegisterLibraryServiceHandler(ctx, gwMux, conn))
				return retErr
			},
		)
	})

	// checks if we have some errors from the go routines
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
}

// checks if we have any error. If so then we
func check(err error, msg string) {
	structuredLogger, _ := zap.NewProduction()
	log := structuredLogger.Sugar()
	if err != nil {
		log.Infow("%v, err: %v\n", msg, err)
		os.Exit(1)
	}
}
