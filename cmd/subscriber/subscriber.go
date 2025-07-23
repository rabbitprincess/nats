package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	nats_go "github.com/nats-io/nats.go"
	"github.com/rabbitprincess/nats/nats"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "subscriber",
		Short: "nats subscriber",
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	host    string
	port    int
	subject string
)

func main() {
	fs := cmd.PersistentFlags()
	fs.StringVarP(&host, "host", "H", "localhost", "NATS server host")
	fs.IntVarP(&port, "port", "P", 8000, "NATS server port")
	fs.StringVarP(&subject, "subject", "s", "default", "NATS subject")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func runServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs
		cancel()
	}()

	serv := nats.NewServer(host, port)
	if err := serv.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start NATS server")
	}

	cli := nats.NewClient(host, port)
	if err := cli.Connect(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}

	defer cli.Drain()
	_, err := cli.Conn.Subscribe(subject, func(m *nats_go.Msg) {
		log.Info().Msgf("Received on [%s]: %s", m.Subject, string(m.Data))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe")
	}

	<-ctx.Done()

	log.Info().Msg("NATS subscriber is running")
}
