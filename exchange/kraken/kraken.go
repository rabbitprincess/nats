package kraken

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thrasher-corp/gocryptotrader/config"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kraken"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
)

type Kraken struct {
	cfg *config.Exchange
	ex  *kraken.Exchange
}

func NewKraken() (*Kraken, error) {
	kraken := new(kraken.Exchange)
	cfg, err := exchange.GetDefaultConfig(context.Background(), kraken)
	if err != nil {
		return nil, err
	}

	// configure custom settings if needed
	// cfg.Enabled = true

	if err := kraken.Setup(cfg); err != nil {
		return nil, err
	}
	if err := kraken.Websocket.Enable(); err != nil {
		return nil, err
	}

	return &Kraken{
		cfg: cfg,
		ex:  kraken,
	}, nil
}

func (k *Kraken) GetExchange() exchange.IBotExchange {
	return k.ex
}

func (k *Kraken) Subscribe(sub ...*subscription.Subscription) (chan any, error) {
	err := k.ex.Websocket.SubscribeToChannels(k.ex.Websocket.Conn, sub)
	if err != nil {
		return nil, err
	}

	return k.ex.Websocket.DataHandler, nil
}

func (k *Kraken) SubscribeFunc(ctx context.Context, fn func(data any), sub ...*subscription.Subscription) error {
	dataChan, err := k.Subscribe(sub...)
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
				if !k.ex.Websocket.IsConnected() && !k.ex.Websocket.IsConnecting() {
					if err := k.ex.Websocket.Connect(); err != nil {
						log.Error().Err(err).Msg("❌ Failed to reconnect websocket")
					}
					if err := k.ex.Websocket.SubscribeToChannels(k.ex.Websocket.Conn, sub); err != nil {
						log.Error().Err(err).Msg("❌ Failed to resubscribe after reconnect")
					}
				}
			}
		}
	}()
	return nil
}
