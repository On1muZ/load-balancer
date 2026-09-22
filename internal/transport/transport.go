package transport

import (
	"net"
	"net/http"
	"time"
)

func NewTransport() http.RoundTripper {

	return &http.Transport{

		// Сколько соединений держим прогретыми
		MaxIdleConns:        50000,
		MaxIdleConnsPerHost: 10000,

		// КРИТИЧЕСКИ ВАЖНО:
		// Если MaxConnsPerHost не равен 0, запросы жестко блокируются в очереди!
		MaxConnsPerHost: 0,

		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,

		// Ускоряем установку TCP-хендшейка
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
}
