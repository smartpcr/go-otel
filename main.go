package main

import (
	"context"
	"fmt"
	"github.com/smartpcr/go-otel/pkg/ot"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	// Run until signaled or the context expires.
	go func() {
		select {
		case <-c:
			fmt.Println("CTRL-C")
			gracefulStop()
			cancel()
		case <-ctx.Done():
			gracefulStop()
		}
	}()

	fmt.Println("registering logger")
	logger := ot.RegisterLogger(ctx)
	logger.Infof("starting %s", ServiceName)

	logger.Infof("registering tracing at %s", config.Receiver.Endpoint)
	err := ot.RegisterTracing(ctx, config.Receiver.Endpoint, ServiceName, logger)
	if err != nil {
		panic(err)
	}

	ctx, span, startupLogger := ot.StartSpanLogger(ctx)
	defer span.End()
	span.AddEvent("test event")
	startupLogger.Infof("test log")

	logger.Infof("registering metrics at %s", config.Receiver.Endpoint)
	metric, err := ot.RegisterOtelMetrics(ctx, config.Receiver.Endpoint, ServiceName)
	if err != nil {
		panic(err)
	}

	counter, err := metric.Meter("test").Int64Counter("test.counter")
	if err != nil {
		panic(err)
	}
	counter.Add(ctx, 1)
}

func gracefulStop() {
	fmt.Println("graceful stop")
}
