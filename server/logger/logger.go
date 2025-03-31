package logger

import (
	"context"
	"os"

	"github.com/vedadiyan/lokiclient"
)

var LokiClient *lokiclient.Client

func InitLogger() *lokiclient.Client {
	lokiURL := os.Getenv("LOKI_URL")
	if lokiURL == "" {
		lokiURL = "http://loki:3100/loki/api/v1/push"
	}
    LokiClient = lokiclient.NewClient(lokiURL)
    
    stream := lokiclient.NewStreamCustom(map[string]string{
        "app": "slow-server",
        "environment": "production",
    })
    
    LokiClient.Info(context.Background(), stream, "Logger initialized")
    
    return LokiClient
}   


func GetLogger() *lokiclient.Client {
    return LokiClient
}
