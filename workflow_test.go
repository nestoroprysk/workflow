package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	Name    string
	Surname string
}

type NameReader struct {
	resp http.ResponseWriter
	req  *http.Request
}

type SurnameReader struct {
	resp http.ResponseWriter
	req  *http.Request
}

type StateLS struct {
	State
}

func (sls *StateLS) Load() (State, error) {
	return sls.State, nil
}

func (sls *StateLS) Save(s State) error {
	sls.State = s
	return nil
}

func NewNameReader(resp http.ResponseWriter, req *http.Request) *NameReader {
	return &NameReader{
		resp: resp,
		req:  req,
	}
}

func (nr *NameReader) Run(ctx context.Context, csd *StateDecoder) (any, error) {
	var result Person
	err := json.NewDecoder(nr.req.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func NewSurnameReader(resp http.ResponseWriter, req *http.Request) *SurnameReader {
	return &SurnameReader{
		resp: resp,
		req:  req,
	}
}

func (nr *SurnameReader) Run(ctx context.Context, csd *StateDecoder) (any, error) {
	var p Person
	err := json.NewDecoder(nr.req.Body).Decode(&p)
	if err != nil {
		return nil, err
	}

	var result Person
	err = csd.Decode(&result)
	if err != nil {
		return nil, err
	}

	result.Surname = p.Surname

	return result, nil
}

func TestWorkflow(t *testing.T) {
	w := NewWorkflow(
		&StateLS{},
		NewNameReader(
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodPost, "/name", strings.NewReader(`{"name": "Peter"}`)),
		),
		NewSurnameReader(
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodPost, "/surname", strings.NewReader(`{"surname": "Parker"}`)),
		),
	)

	result, err := w.Continue(context.TODO())
	require.NoError(t, err)
	require.Nil(t, result, "should not return result yet because the workflow is not ower")

	result, err = w.Continue(context.TODO())
	require.NoError(t, err)
	require.NotNil(t, result, "should return result because the workflow is ower")

	var p Person
	require.NoError(t, result.Decode(&p))
	assert.Equal(t, "Peter", p.Name)
	assert.Equal(t, "Parker", p.Surname)
}
