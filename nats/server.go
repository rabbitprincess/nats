package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/rs/zerolog/log"
)

type EmbeddedServer struct {
	Server *server.Server
	Host   string
	Port   int
}

func NewServer(host string, port int) *EmbeddedServer {
	return &EmbeddedServer{
		Host: host,
		Port: port,
	}
}

func (s *EmbeddedServer) Address() string {
	return fmt.Sprintf("nats://%s:%d", s.Host, s.Port)
}

func (s *EmbeddedServer) Start(ctx context.Context) error {
	var err error
	s.Server, err = server.NewServer(&server.Options{
		Host: s.Host,
		Port: s.Port,
	})
	if err != nil {
		return err
	}
	go s.Server.Start()
	if s.Server.ReadyForConnections(5 * time.Second) {
		log.Info().Msgf("NATS server ready at %s...", s.Address())
	}

	log.Info().Msgf("NATS server started at %s", s.Address())
	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down NATS server...")
		s.Server.Shutdown()
	}()

	return nil
}
