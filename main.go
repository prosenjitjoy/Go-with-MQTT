package main

import (
	"crypto/tls"
	"flag"
	"log"
	"main/hooks"
	"os"
	"os/signal"
	"syscall"

	badgerdb "github.com/dgraph-io/badger/v4"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/hooks/storage/badger"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func main() {
	certFile := flag.String("cert", "server.crt", "TLS certificate file")
	keyFile := flag.String("key", "server.key", "TLS key file")
	flag.Parse()

	badgerPath := ".badger"
	defer os.RemoveAll(badgerPath)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	err = server.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		log.Fatal(err)
	}

	badgerOpts := badgerdb.DefaultOptions(badgerPath)
	badgerOpts.ValueLogFileSize = 100 * (1 << 20)

	err = server.AddHook(new(badger.Hook), &badger.Options{
		Path:           badgerPath,
		GcInterval:     5 * 60,
		GcDiscardRatio: 0.5,
		Options:        &badgerOpts,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = server.AddHook(new(hooks.ExampleHook), &hooks.ExampleHookOptions{
		Server: server,
	})

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: ":1883",
	})

	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	tls := listeners.NewTCP(listeners.Config{
		ID:        "tls",
		Address:   ":8883",
		TLSConfig: tlsConfig,
	})

	err = server.AddListener(tls)
	if err != nil {
		log.Fatal(err)
	}

	ws := listeners.NewWebsocket(listeners.Config{
		ID:        "wss",
		Address:   ":8884",
		TLSConfig: tlsConfig,
	})

	err = server.AddListener(ws)
	if err != nil {
		log.Fatal(err)
	}

	stats := listeners.NewHTTPStats(
		listeners.Config{
			ID:        "stats",
			Address:   ":8080",
			TLSConfig: tlsConfig,
		},
		server.Info,
	)

	err = server.AddListener(stats)
	if err != nil {
		log.Fatal(err)
	}

	hc := listeners.NewHTTPHealthCheck(listeners.Config{
		ID:        "health",
		Address:   ":8081",
		TLSConfig: tlsConfig,
	})

	err = server.AddListener(hc)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn("caught signal, stopping...")
	err = server.Close()
	if err != nil {
		log.Fatal(err)
	}
	server.Log.Info("main.go finished")
}
