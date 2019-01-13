package codefresh

import (
	"encoding/json"
	"time"
)

type (
	ITokenAPI interface {
		Create(name string, subject string) (*Token, error)
		List() ([]*Token, error)
	}

	Token struct {
		ID          string    `json:"_id"`
		Name        string    `json:"name"`
		TokenPrefix string    `json:"tokenPrefix"`
		Created     time.Time `json:"created"`
		Subject     struct {
			Type string `json:"type"`
			Ref  string `json:"ref"`
		} `json:"subject"`
		Value string
	}

	tokenSubjectType int

	getTokensReponse struct {
		Tokens []*Token
	}

	token struct {
		codefresh Codefresh
	}
)

const (
	runtimeEnvironmentSubject tokenSubjectType = 0
)

func newTokenAPI(codefresh Codefresh) ITokenAPI {
	return &token{codefresh}
}

func (s tokenSubjectType) String() string {
	return [...]string{"runtime-environment"}[s]
}

func (t *token) Create(name string, subject string) (*Token, error) {
	resp, err := t.codefresh.requestAPI(&requestOptions{
		path:   "/api/auth/key",
		method: "POST",
		body: map[string]interface{}{
			"name": name,
		},
		qs: map[string]string{
			"subjectReference": subject,
			"subjectType":      runtimeEnvironmentSubject.String(),
		},
	})
	value, err := t.codefresh.getBodyAsString(resp)
	if err != nil {
		return nil, err
	}
	return &Token{
		Name:  name,
		Value: value,
	}, err
}

func (t *token) List() ([]*Token, error) {
	emptySlice := make([]*Token, 0)
	resp, err := t.codefresh.requestAPI(&requestOptions{
		path:   "/api/auth/keys",
		method: "GET",
	})
	tokensAsBytes, err := t.codefresh.getBodyAsBytes(resp)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(tokensAsBytes, &emptySlice)

	return emptySlice, err
}
