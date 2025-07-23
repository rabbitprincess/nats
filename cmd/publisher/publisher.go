package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rabbitprincess/nats/nats"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "publisher",
		Short: "nats publisher",
		Run: func(cmd *cobra.Command, args []string) {
			runPublisher()
		},
	}

	host    string
	port    int
	subHost string
	subPort int
	subject string
)

func main() {
	fs := cmd.PersistentFlags()
	fs.StringVarP(&host, "host", "H", "localhost", "NATS publisher host")
	fs.IntVarP(&port, "port", "P", 9000, "NATS publisher port")
	fs.StringVar(&subHost, "subhost", "localhost", "NATS subscriber host")
	fs.IntVar(&subPort, "subport", 8000, "NATS subscriber port")
	fs.StringVarP(&subject, "subject", "s", "default", "NATS subject")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func runPublisher() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs
		cancel()
	}()

	// Connect to NATS
	cli := nats.NewClient(host, port)
	if err := cli.Connect(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer cli.Drain()

	log.Info().Msgf("Publishing to subject [%s]", subject)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> Enter message (or type 'exit'): ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("Error reading input")
			continue
		}

		text = strings.TrimSpace(text)
		if text == "exit" {
			log.Info().Msg("Exiting publisher...")
			break
		}
		if text == "" {
			continue
		}

		if err := cli.Conn.Publish(subject, []byte(text)); err != nil {
			log.Error().Err(err).Msg("Failed to publish message")
		} else {
			log.Info().Msgf("Published: %s", text)
		}
	}

	<-ctx.Done()
	time.Sleep(1 * time.Second)
}
