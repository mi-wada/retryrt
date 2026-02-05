package retryrt

import (
	"math"
	"math/rand"
	"net/http"
	"time"
)

var _ http.RoundTripper = (*roundTripper)(nil)

type roundTripper struct {
	base       http.RoundTripper
	maxRetries int
	backoff    Backoff
}

var (
	defaultMaxRetries = 3
	defaultBackoffMin = 1 * time.Second
	defaultBackoffMax = 30 * time.Second
)

func New(base http.RoundTripper, opts ...Option) *roundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	rt := &roundTripper{
		base:       base,
		maxRetries: defaultMaxRetries,
		backoff:    DefaultBackoff(defaultBackoffMin, defaultBackoffMax),
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

// RoundTrip implements the [http.RoundTripper] interface.
func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request.
	// https://cs.opensource.google/go/go/+/refs/tags/go1.25.1:src/net/http/client.go;l=128-132
	clonedReq := req.Clone(req.Context())
	var resp *http.Response
	var err error
	for i := 0; i < rt.maxRetries+1; i++ {
		resp, err = rt.base.RoundTrip(clonedReq)
		if shouldRetry(resp, err) {
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(rt.backoff(i, resp))
			continue
		}
		break
	}
	return resp, err
}

type Option func(*roundTripper)

// WithMaxRetries sets the maximum number of retries for a request.
func WithMaxRetries(n int) Option {
	return func(rt *roundTripper) {
		rt.maxRetries = n
	}
}

// WithBackoff sets the backoff strategy for retries.
func WithBackoff(b Backoff) Option {
	return func(rt *roundTripper) {
		rt.backoff = b
	}
}

type Backoff func(attemptNum int, resp *http.Response) time.Duration

func DefaultBackoff(min, max time.Duration) Backoff {
	return func(attemptNum int, resp *http.Response) time.Duration {
		mult := math.Pow(2, float64(attemptNum))
		wait := time.Duration(float64(min) * mult)
		if wait > max || wait < min {
			wait = max
		}
		if wait > 0 {
			wait = time.Duration(rand.Int63n(int64(wait)))
		}
		return wait
	}
}

// TODO: implement. リトライカウントも渡すべき？あとリクエストも渡すべきかな？そしてこのshouldRetryはDefaultShouldRetryとして公開すべきなんだろうな。その上でWithShouldRetryみたいなオプションを提供してカスタマイズできるようにする。
func shouldRetry(resp *http.Response, err error) bool {
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return false
	}
	return true
}
