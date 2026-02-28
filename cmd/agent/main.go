package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/htrandev/metrics/internal/agent"
	metricsclient "github.com/htrandev/metrics/internal/agent/client"
	"github.com/htrandev/metrics/internal/config"
	"github.com/htrandev/metrics/internal/info"
	"github.com/htrandev/metrics/internal/model"
	"github.com/htrandev/metrics/internal/proto"
	"github.com/htrandev/metrics/pkg/crypto"
	"github.com/htrandev/metrics/pkg/logger"
	"github.com/htrandev/metrics/pkg/netutil"
)

func main() {
	if err := run(); err != nil {
		log.Printf("run agent ends with error: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	info.PrintBuildInfo()

	log.Println("init config")
	conf, err := config.GetAgentConfig()
	if err != nil {
		return fmt.Errorf("get agent config: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(conf.LogLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	zl.Info("init collection")
	collection := model.NewCollection()

	zl.Info("init public key")
	publicKey, err := crypto.PublicKey(conf.PublicKeyFile)
	if err != nil {
		return fmt.Errorf("init public key: %w", err)
	}

	ip, err := netutil.GetLocalIP()
	if err != nil {
		return fmt.Errorf("get local ip addr: %w", err)
	}

	var client agent.Client
	if conf.UseGRPC {
		zl.Info("init grpc conn")
		conn, err := grpc.NewClient(conf.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("init grpc client conn: %w", err)
		}
		defer conn.Close()
		zl.Info("init grpc client")
		grpcClient := proto.NewMetricsClient(conn)

		zl.Info("create grpc client")
		client = metricsclient.NewGRPC(grpcClient,
			metricsclient.WithMaxRetry(conf.MaxRetry),
			metricsclient.WithAddr(conf.Addr),
			metricsclient.WithLogger(zl),
			metricsclient.WithPublicKey(publicKey),
			metricsclient.WithSignature(conf.Signature),
			metricsclient.WithIP(ip.String()),
		)
	} else {
		zl.Info("init resty client")
		restyClient := resty.New().
			SetTimeout(30 * time.Second)

		zl.Info("create http client")
		client = metricsclient.NewHTTP(restyClient,
			metricsclient.WithMaxRetry(conf.MaxRetry),
			metricsclient.WithAddr(conf.Addr),
			metricsclient.WithLogger(zl),
			metricsclient.WithPublicKey(publicKey),
			metricsclient.WithSignature(conf.Signature),
			metricsclient.WithIP(ip.String()),
		)
	}

	zl.Info("init agent")
	agent := agent.New(&agent.AgentOptions{
		Logger:         zl,
		Client:         client,
		Collector:      collection,
		RateLimit:      conf.RateLimit,
		PollInterval:   conf.PollInterval,
		ReportInterval: conf.ReportInterval,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	agent.Run(ctx)
	return nil
}
