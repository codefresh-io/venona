package codefresh

import (
	gentleman "gopkg.in/h2non/gentleman.v2"
)

type (
	// AuthOptions
	AuthOptions struct {

		// Token - Codefresh token
		Token string
	}

	// Options
	ClientOptions struct {
		Auth AuthOptions

		Host string
	}

	codefresh struct {
		token  string
		client *gentleman.Client
	}

	requestOptions struct {
		path   string
		method string
		body   map[string]interface{}
		qs     map[string]string
	}
)
