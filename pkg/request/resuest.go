package request

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

// Option defines a function type that applies a configuration to the resty.Client.
type Option func(*resty.Client)

// Client represents the HTTP client with configured options and context.
type Client struct {
	// RetryCount defines the number of retry attempts for failed requests.
	RetryCount     int
	client         *resty.Client
	requestOptions []func(*resty.Request)
	ctx            context.Context
}

// MustParseDuration parses a duration string and panics if parsing fails.
func MustParseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		panic("util: Can't parse duration `" + durationStr + "`: " + err.Error())
	}
	return duration
}

// WithContext sets the context for the Client.
func (c *Client) WithContext(ctx context.Context) *Client {
	c.ctx = ctx
	return c
}

// EnableTrace enables tracing for the resty client.
func (c *Client) EnableTrace() Option {
	return func(client *resty.Client) {
		client.EnableTrace()
	}
}

// Timeout sets the request timeout duration.
func Timeout(duration string) Option {
	return func(client *resty.Client) {
		client.SetTimeout(MustParseDuration(duration))
	}
}

// BaseURL sets the base URL for the resty client.
func BaseURL(url string) Option {
	return func(client *resty.Client) {
		client.SetBaseURL(url)
	}
}

// Header sets the headers for the resty client.
func Header(headers map[string]string) Option {
	return func(client *resty.Client) {
		client.SetHeader("Content-Type", "application/json")
		for key, value := range headers {
			client.SetHeader(key, value)
		}
	}
}

// Query sets the query parameters for the resty client.
func Query(params map[string]string) Option {
	return func(client *resty.Client) {
		client.SetQueryParams(params)
	}
}

// SetErrorHandler sets the error handler for the resty client.
func SetErrorHandler(errHandler interface{}) Option {
	return func(client *resty.Client) {
		client.SetError(errHandler)
	}
}

// AuthToken sets the authentication token for the resty client.
func AuthToken(token string) Option {
	return func(client *resty.Client) {
		client.SetAuthToken(token)
	}
}

// Logger sets a custom logger for the resty client.
func Logger(customLogger *zap.SugaredLogger) Option {
	return func(client *resty.Client) {
		client.SetLogger(customLogger)
	}
}

// SetRetryCount configures the number of retry attempts.
func SetRetryCount(count int) Option {
	return func(client *resty.Client) {
		client.SetRetryCount(count)
	}
}

// SetPathParams sets the path parameters for the resty client.
func (c *Client) SetPathParams(params map[string]string) Option {
	return func(client *resty.Client) {
		client.SetPathParams(params)
	}
}

// SetBody sets the request body.
func (c *Client) SetBody(body interface{}) *Client {
	c.requestOptions = append(c.requestOptions, func(req *resty.Request) {
		req.SetBody(body)
	})
	return c
}

// SetFormData sets the form data for the request.
func (c *Client) SetFormData(formData map[string]string) *Client {
	c.requestOptions = append(c.requestOptions, func(req *resty.Request) {
		req.SetFormData(formData)
	})
	return c
}

// SetResult sets the result object to store the response.
func (c *Client) SetResult(result interface{}) *Client {
	c.requestOptions = append(c.requestOptions, func(req *resty.Request) {
		req.SetResult(result)
	})
	return c
}

// SetError sets the error object to store error responses.
func (c *Client) SetError(err interface{}) *Client {
	c.requestOptions = append(c.requestOptions, func(req *resty.Request) {
		req.SetError(err)
	})
	return c
}

// NewClient initializes and returns a new Client with applied options.
func NewClient(options ...Option) *Client {
	c := &Client{
		client: resty.New(),
	}

	// Configure JSON marshaller and unmarshaller
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	c.client.JSONMarshal = json.Marshal
	c.client.JSONUnmarshal = json.Unmarshal

	// Set default timeout to 18 seconds
	c.client.SetTimeout(MustParseDuration("18s"))
	c.client.SetCloseConnection(true)

	// Configure retry settings
	c.client.
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(time.Minute)

	// Apply provided options
	for _, option := range options {
		option(c.client)
	}

	return c
}

// Response represents the HTTP response with status code and data.
type Response struct {
	StatusCode int
	Data       []byte
}

// Do sends an HTTP request with the specified method and URL.
func (c *Client) Do(method string, url string) (*Response, error) {
	var (
		res *resty.Response
		err error
	)
	req := c.client.R()

	// Inject tracing headers if context is set
	if c.ctx != nil {
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(c.ctx, propagation.HeaderCarrier(req.Header))
	}

	req.SetContentLength(true)
	for _, opt := range c.requestOptions {
		opt(req)
	}

	// Execute the request based on the method
	switch method {
	case "GET":
		res, err = req.Get(url)
	case "POST":
		res, err = req.Post(url)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: res.StatusCode(),
		Data:       res.Body(),
	}, nil
}
