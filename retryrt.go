package retryrt

import "net/http"

var _ http.RoundTripper = (*roundTripper)(nil)

type roundTripper struct {
	base     http.RoundTripper
	retryMax int
}

type Option func(*roundTripper)

// WithRetryMax sets the maximum number of retries for a request.
func WithRetryMax(n int) Option {
	return func(rt *roundTripper) {
		rt.retryMax = n
	}
}

func New(base http.RoundTripper, opts ...Option) *roundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	rt := &roundTripper{
		base: base,
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request.
	// https://cs.opensource.google/go/go/+/refs/tags/go1.25.1:src/net/http/client.go;l=128-132
	clonedReq := req.Clone(req.Context())
	var resp *http.Response
	var err error
	for i := 0; i < rt.retryMax+1; i++ {
		resp, err = rt.base.RoundTrip(clonedReq)
		if shouldRetry(resp, err) {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}
		break
	}
	return resp, err
}

// TODO: implement. リトライカウントも渡すべき？あとリクエストも渡すべきかな？そしてこのshouldRetryはDefaultShouldRetryとして公開すべきなんだろうな。その上でWithShouldRetryみたいなオプションを提供してカスタマイズできるようにする。
func shouldRetry(resp *http.Response, err error) bool {
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return false
	}
	return true
}
