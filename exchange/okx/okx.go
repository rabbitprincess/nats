package okx

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thrasher-corp/gocryptotrader/config"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/okx"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
)

type Okx struct {
	cfg *config.Exchange
	ex  *okx.Exchange
}

func NewOkx() (*Okx, error) {
	okx := new(okx.Exchange)
	cfg, err := exchange.GetDefaultConfig(context.Background(), okx)
	if err != nil {
		return nil, err
	}

	// configure custom settings if needed
	// cfg.Enabled = true

	if err := okx.Setup(cfg); err != nil {
		return nil, err
	}
	if err := okx.Websocket.Enable(); err != nil {
		return nil, err
	}

	return &Okx{
		cfg: cfg,
		ex:  okx,
	}, nil
}

func (o *Okx) GetExchange() exchange.IBotExchange {
	return o.ex
}

func (o *Okx) Subscribe(sub ...*subscription.Subscription) (chan any, error) {
	err := o.ex.Websocket.SubscribeToChannels(o.ex.Websocket.Conn, sub)
	if err != nil {
		return nil, err
	}

	return o.ex.Websocket.DataHandler, nil
}

func (o *Okx) SubscribeFunc(ctx context.Context, fn func(data any), sub ...*subscription.Subscription) error {
	dataChan, err := o.Subscribe(sub...)
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
				if !o.ex.Websocket.IsConnected() && !o.ex.Websocket.IsConnecting() {
					if err := o.ex.Websocket.Connect(); err != nil {
						log.Error().Err(err).Msg("❌ Failed to reconnect websocket")
					}
					if err := o.ex.Websocket.SubscribeToChannels(o.ex.Websocket.Conn, sub); err != nil {
						log.Error().Err(err).Msg("❌ Failed to resubscribe after reconnect")
					}
				}
			}
		}
	}()
	return nil
}
