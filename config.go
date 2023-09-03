package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

const ServiceName = "open-telemetry-test"

type Receiver struct {
	Endpoint string `yaml:"endpoint"`
}

type Config struct {
	Receiver Receiver `yaml:"receiver"`
}

var config Config

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
	})

	//// Set default values
	//viper.SetDefault("loki.endpoint", "http://loki.local:3100/api/prom/push")
	//viper.SetDefault("prometheus.endpoint", "http://prometheus.local:9090")
	//viper.SetDefault("jaeger.trace_endpoint", "http://jaeger.local:14250/api/traces")
	//viper.SetDefault("jaeger.prometheus_endpoint", "http://jaeger.local:16686/metrics")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct, %s", err)
	}

	// Print to verify
	fmt.Println("Open Telemetry Receiver Endpoint:", config.Receiver.Endpoint)
}
