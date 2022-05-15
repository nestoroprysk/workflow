package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

// Workflow forwars the call to the proper persisted stage of the workflow.
type Workflow struct {
	StateLoaderSaver
	Stages []StageRunner
}

// StateLoaderSaver reads and persists the workflow state.
type StateLoaderSaver interface {
	Load() (State, error)
	Save(State) error
}

// State is a persisted workflow representation.
type State struct {
	CurrentStage int
	Result       *StateDecoder
}

// StageRunner is a runnable step in a workflow.
type StageRunner interface {
	Run(context.Context, *StateDecoder) (any, error)
}

// StateDecoder decodes the current state input the input argument.
type StateDecoder struct {
	io.Reader
}

// NewStateDecoder provides input for each stage.
func NewStateDecoder(a any) (*StateDecoder, error) {
	result, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return &StateDecoder{
		Reader: bytes.NewBuffer(result),
	}, nil
}

// Decode provides input for a stage.
func (csd *StateDecoder) Decode(a any) error {
	return json.NewDecoder(csd).Decode(a)
}

// NewWorkflow defines a workflow.
func NewWorkflow(sls StateLoaderSaver, stages ...StageRunner) *Workflow {
	return &Workflow{
		StateLoaderSaver: sls,
		Stages:           stages,
	}
}

// Continue proceeds with the workflow.
//
// If the stage errors, nothing is done.
//
// If the workflow is done,
// - not nil result is returned and
// - the next call starts a new workflow.
func (w *Workflow) Continue(ctx context.Context) (*StateDecoder, error) {
	s, err := w.Load()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = w.Save(s)
	}()

	a, err := w.Stages[s.CurrentStage].Run(ctx, s.Result)
	if err != nil {
		return nil, err
	}

	result, err := NewStateDecoder(a)
	if err != nil {
		return nil, err
	}

	lastStage := s.CurrentStage == len(w.Stages)-1
	if lastStage {
		s = State{}
		return result, nil
	}

	s.Result = result
	s.CurrentStage++

	return nil, nil
}
