// Package webhook announces releases via HTTP POST requests.
package webhook

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/internal/tmpl"
	"github.com/goreleaser/goreleaser/v2/pkg/context"
)

const (
	defaultMessageTemplate = `{ "message": "{{ .ProjectName }} {{ .Tag }} is out! Check it out at {{ .ReleaseURL }}"}`
	contentTypeHeaderKey   = "Content-Type"
	userAgentHeaderKey     = "User-Agent"
	userAgentHeaderValue   = "goreleaser"
	authorizationHeaderKey = "Authorization"
	defaultContentType     = "application/json; charset=utf-8"
)

var defaultExpectedStatusCodes = []int{
	http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent,
}

// Pipe implementation.
type Pipe struct{}

func (Pipe) String() string { return "webhook" }

// Skip implements Skipper.
func (Pipe) Skip(ctx *context.Context) (bool, error) {
	enable, err := tmpl.New(ctx).Bool(ctx.Config.Announce.Webhook.Enabled)
	return !enable, err
}

type envConfig struct {
	BasicAuthHeader   string `env:"BASIC_AUTH_HEADER_VALUE"`
	BearerTokenHeader string `env:"BEARER_TOKEN_HEADER_VALUE"`
}

// Default sets the pipe defaults.
func (p Pipe) Default(ctx *context.Context) error {
	if ctx.Config.Announce.Webhook.MessageTemplate == "" {
		ctx.Config.Announce.Webhook.MessageTemplate = defaultMessageTemplate
	}
	if ctx.Config.Announce.Webhook.ContentType == "" {
		ctx.Config.Announce.Webhook.ContentType = defaultContentType
	}
	if len(ctx.Config.Announce.Webhook.ExpectedStatusCodes) == 0 {
		ctx.Config.Announce.Webhook.ExpectedStatusCodes = defaultExpectedStatusCodes
	}
	return nil
}

// Announce implements Announcer.
func (p Pipe) Announce(ctx *context.Context) error {
	cfg, err := env.ParseAs[envConfig]()
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	endpointURLConfig, err := tmpl.New(ctx).Apply(ctx.Config.Announce.Webhook.EndpointURL)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}
	if len(endpointURLConfig) == 0 {
		return errors.New("webhook: no endpoint url")
	}

	if _, err := url.ParseRequestURI(endpointURLConfig); err != nil {
		return fmt.Errorf("webhook: %w", err)
	}
	endpointURL, err := url.Parse(endpointURLConfig)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	msg, err := tmpl.New(ctx).Apply(ctx.Config.Announce.Webhook.MessageTemplate)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}

	log.Infof("posting: '%s'", msg)
	customTransport := http.DefaultTransport.(*http.Transport).Clone()

	customTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: ctx.Config.Announce.Webhook.SkipTLSVerify,
	}

	client := &http.Client{
		Transport: customTransport,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), strings.NewReader(msg))
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}
	req.Header.Add(contentTypeHeaderKey, ctx.Config.Announce.Webhook.ContentType)
	req.Header.Add(userAgentHeaderKey, userAgentHeaderValue)

	if cfg.BasicAuthHeader != "" {
		log.Debugf("set basic auth header")
		req.Header.Add(authorizationHeaderKey, cfg.BasicAuthHeader)
	} else if cfg.BearerTokenHeader != "" {
		log.Debugf("set bearer token header")
		req.Header.Add(authorizationHeaderKey, cfg.BearerTokenHeader)
	}

	for key, value := range ctx.Config.Announce.Webhook.Headers {
		log.Debugf("Header Key %s / Value %s", key, value)
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: %w", err)
	}
	defer resp.Body.Close()

	if !slices.Contains(ctx.Config.Announce.Webhook.ExpectedStatusCodes, resp.StatusCode) {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("request failed with status %v", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	log.Infof("Post OK: '%v'", resp.StatusCode)
	log.Infof("Response : %v\n", string(body))
	return nil
}
