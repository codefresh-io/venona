package codefresh

import (
	"fmt"
	"net/url"

	"github.com/codefresh-io/go-sdk/internal"

	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/plugins/body"
	"gopkg.in/h2non/gentleman.v2/plugins/query"
)

type (
	Codefresh interface {
		requestAPI(*requestOptions) (*gentleman.Response, error)
		ITokenAPI
		IPipelineAPI
		IRuntimeEnvironmentAPI
	}
)

func New(opt *ClientOptions) Codefresh {
	client := gentleman.New()
	client.URL(opt.Host)
	return &codefresh{
		token:  opt.Auth.Token,
		client: client,
	}
}

func (c *codefresh) requestAPI(opt *requestOptions) (*gentleman.Response, error) {
	req := c.client.Request()
	url, err := url.Parse(opt.path)
	internal.DieOnError(err)
	req.Path(url.String())
	req.Method(opt.method)
	req.AddHeader("Authorization", c.token)
	if opt.body != nil {
		req.Use(body.JSON(opt.body))
	}
	if opt.qs != nil {
		for k, v := range opt.qs {
			req.Use(query.Set(k, v))
		}
	}
	res, _ := req.Send()
	if res.StatusCode > 400 {
		return res, fmt.Errorf("Error occured during API invocation\nError: %s", res.String())
	}
	return res, nil
}
