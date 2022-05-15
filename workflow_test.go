package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	Name    string
	Surname string
}

type NameReader struct {
	Name string
}

type SurnameReader struct {
	Surname string
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

func (nr *NameReader) Run(ctx context.Context, csd *StateDecoder) (any, error) {
	return Person{Name: nr.Name}, nil
}

func (sr *SurnameReader) Run(ctx context.Context, csd *StateDecoder) (any, error) {
	var result Person
	err := csd.Decode(&result)
	if err != nil {
		return nil, err
	}

	result.Surname = sr.Surname

	return result, nil
}

func TestWorkflow(t *testing.T) {
	nr := &NameReader{Name: "Peter"}
	sr := &SurnameReader{Surname: "Parker"}
	w := NewWorkflow(&StateLS{}, nr, sr)

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

	nr.Name = "Harry"
	sr.Surname = "Potter"

	result, err = w.Continue(context.TODO())
	require.NoError(t, err)
	require.Nil(t, result, "should return nil because the new workflow is pending")

	result, err = w.Continue(context.TODO())
	require.NoError(t, err)
	require.NotNil(t, result, "should return result because the new workflow is ower")

	require.NoError(t, result.Decode(&p))
	assert.Equal(t, "Harry", p.Name)
	assert.Equal(t, "Potter", p.Surname)
}
