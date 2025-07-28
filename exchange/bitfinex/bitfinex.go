package bitfinex

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thrasher-corp/gocryptotrader/config"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/bitfinex"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
)

type Bitfinex struct {
	cfg *config.Exchange
	ex  *bitfinex.Exchange
}

func NewBitfinex() (*Bitfinex, error) {
	ex := new(bitfinex.Exchange)
	cfg, err := exchange.GetDefaultConfig(context.Background(), ex)
	if err != nil {
		return nil, err
	}

	// configure custom settings if needed
	// cfg.Enabled = true

	if err := ex.Setup(cfg); err != nil {
		return nil, err
	}
	if err := ex.Websocket.Enable(); err != nil {
		return nil, err
	}

	return &Bitfinex{
		cfg: cfg,
		ex:  ex,
	}, nil
}

func (b *Bitfinex) GetExchange() exchange.IBotExchange {
	return b.ex
}

func (b *Bitfinex) Subscribe(sub ...*subscription.Subscription) (chan any, error) {
	err := b.ex.Websocket.SubscribeToChannels(b.ex.Websocket.Conn, sub)
	if err != nil {
		return nil, err
	}

	return b.ex.Websocket.DataHandler, nil
}

func (b *Bitfinex) SubscribeFunc(ctx context.Context, fn func(data any), sub ...*subscription.Subscription) error {
	dataChan, err := b.Subscribe(sub...)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Err(ctx.Err()).Msg("🛑 Context done, stopping subscription")
				return
			case msg := <-dataChan:
				log.Info().Msgf("✅ Received message: %T %+v", msg, msg)
				fn(msg)
			case <-time.After(30 * time.Second):
				fmt.Println("⏰ No data in 30s")
				if !b.ex.Websocket.IsConnected() && !b.ex.Websocket.IsConnecting() {
					if err := b.ex.Websocket.Connect(); err != nil {
						log.Error().Err(err).Msg("❌ Failed to reconnect websocket")
					}
					if err := b.ex.Websocket.SubscribeToChannels(b.ex.Websocket.Conn, sub); err != nil {
						log.Error().Err(err).Msg("❌ Failed to resubscribe after reconnect")
					}
				}
			}
		}
	}()
	return nil
}
