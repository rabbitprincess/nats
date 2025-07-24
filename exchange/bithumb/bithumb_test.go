package bithumb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thrasher-corp/gocryptotrader/currency"
	"github.com/thrasher-corp/gocryptotrader/exchange/websocket"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/subscription"
)

func TestGetTicker(t *testing.T) {
	ex, err := NewBithumb()
	require.NoError(t, err, "Failed to create instance")

	tickers, err := ex.ex.GetAllTickers(context.Background())
	require.NoError(t, err, "Failed to get ticker")
	for _, ticker := range tickers {
		fmt.Println("Ticker:", ticker)
	}

}

func TestCall(t *testing.T) {

}

const (
	QualifiedChannelTransaction = `{"type":"transaction","symbols":["BTC_KRW","ETH_KRW"]}`
	QualifiedChannelOrderbook   = `{"type":"orderbookdepth","symbols":["BTC_KRW","ETH_KRW"]}`
	QualifiedChannelTicker      = `{"type":"ticker","symbols":["BTC_KRW","ETH_KRW"],"tickTypes":["30M"]}`
)

func TestSubscription(t *testing.T) {
	ex, err := NewBithumb()
	require.NoError(t, err, "Failed to create instance")

	sub := &subscription.Subscription{
		Enabled:          true,
		Channel:          subscription.TickerChannel,
		QualifiedChannel: QualifiedChannelTicker,
		Asset:            asset.Spot,
		Pairs:            []currency.Pair{currency.NewPair(currency.BTC, currency.USDT)},
		Interval:         kline.TenSecond,
		Key:              42,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = ex.SubscribeFunc(ctx, func(data any) {
		switch v := data.(type) {
		case websocket.KlineData:
			fmt.Println(v.Pair, v.Interval, v.OpenPrice, v.ClosePrice)
		case orderbook.Depth:
			fmt.Println(v.Asset())
		}
	}, sub)
	require.NoError(t, err, "Failed to subscribe")

	time.Sleep(30 * time.Hour)
}
