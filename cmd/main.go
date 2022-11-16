package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"opa-webhook/pkg/webhook"

	"github.com/golang/glog"
)

func main() {

	opaWebHookServerParameters := webhook.OpaWebHookServerParameters{}

	// webhook opa parameters
	flag.IntVar(&opaWebHookServerParameters.Port, "port", 8443, "Webhook server port.")
	flag.StringVar(&opaWebHookServerParameters.CertFile, "tlsCertFile", "./ssl/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&opaWebHookServerParameters.KeyFile, "tlsKeyFile", "./ssl/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(opaWebHookServerParameters.CertFile, opaWebHookServerParameters.KeyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	opaWebhookServer := &webhook.OpaWebhookServer{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", opaWebHookServerParameters.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", opaWebhookServer.HandleRoot)
	mux.HandleFunc("/mutate", opaWebhookServer.HandleMutate)
	opaWebhookServer.Server.Handler = mux

	// start opa webhook server in new goroutine
	glog.Infof("Starting Opa Webhook Server......")
	go func() {
		if err := opaWebhookServer.Server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and server opa webhook server: %v", err)
		}
	}()

	// listening OS shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down opa webhook server gracefully...")
	opaWebhookServer.Server.Shutdown(context.Background())
}
