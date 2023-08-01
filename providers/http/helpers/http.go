package helpers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func DoRequest(parentContext context.Context, requestURL string, requestTimeout time.Duration, headers map[string]string, maxResponseSize int64) ([]byte, error) {
	logger := log.With().Str("url", requestURL).Logger()

	logger.Debug().Msg("getting file from server")
	ctx, cancel := context.WithTimeout(parentContext, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build request")
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	logger.Debug().Msgf("server responded with %d", res.StatusCode)
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "expected status 200, but got %d", res.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body")
	}

	return body, nil
}
