package validator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file pins down whether the Validatable interface (`Validate() error`) is honoured for structs that are
// NESTED inside the struct passed to Validate/Unmarshal, at every level of the recursion: pointer fields, value
// fields, deeper levels, slice elements, pointer slice elements and map values, with pointer and value receivers.

const nvNegativeMessage = "value must not be negative"

// nvLeaf implements Validatable with a pointer receiver.
type nvLeaf struct {
	Value int `json:"value"`
}

func (l *nvLeaf) Validate() error {
	if l.Value < 0 {
		return errors.New(nvNegativeMessage)
	}
	return nil
}

// nvLeafByValue implements Validatable with a value receiver.
type nvLeafByValue struct {
	Value int `json:"value"`
}

func (l nvLeafByValue) Validate() error {
	if l.Value < 0 {
		return errors.New(nvNegativeMessage)
	}
	return nil
}

// nvMiddle has no Validate method of its own; it only holds a Validatable child (second nesting level).
type nvMiddle struct {
	Leaf *nvLeaf `json:"leaf"`
}

// Root types WITHOUT their own Validate method, one per nesting shape, so one failing shape cannot hide another.
type (
	nvRootPtr struct {
		Leaf *nvLeaf `json:"leaf"`
	}
	nvRootVal struct {
		Leaf nvLeaf `json:"leaf"`
	}
	nvRootByValue struct {
		Leaf nvLeafByValue `json:"leaf"`
	}
	nvRootDeep struct {
		Middle *nvMiddle `json:"middle"`
	}
	nvRootSlice struct {
		Leaves []nvLeaf `json:"leaves" validate:"dive"`
	}
	nvRootPtrSlice struct {
		Leaves []*nvLeaf `json:"leaves" validate:"dive"`
	}
	nvRootMap struct {
		Leaves map[string]nvLeaf `json:"leaves" validate:"dive"`
	}
)

// nvMessages returns the messages of every field error inside err.
func nvMessages(t *testing.T, err error) []string {
	t.Helper()
	var ve *ValidationError
	require.ErrorAs(t, err, &ve, "want *ValidationError, got %T: %v", err, err)
	msgs := make([]string, 0, len(ve.Errors))
	for i := range ve.Errors {
		msgs = append(msgs, ve.Errors[i].Message)
	}
	return msgs
}

// TestValidatable_Root is the baseline: Validate() on the root struct is honoured.
func TestValidatable_Root(t *testing.T) {
	t.Run("root with an invalid value", func(t *testing.T) {
		err := New[nvLeaf]().Validate(&nvLeaf{Value: -1})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
	})
	t.Run("root with a valid value", func(t *testing.T) {
		require.NoError(t, New[nvLeaf]().Validate(&nvLeaf{Value: 1}))
	})
}

// TestValidatable_NestedValidate calls Validate on a root struct that holds the Validatable struct nested.
func TestValidatable_NestedValidate(t *testing.T) {
	t.Run("pointer field, pointer receiver", func(t *testing.T) {
		err := New[nvRootPtr]().Validate(&nvRootPtr{Leaf: &nvLeaf{Value: -1}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootPtr]().Validate(&nvRootPtr{Leaf: &nvLeaf{Value: 1}}))
	})
	t.Run("nil pointer field is skipped without a panic", func(t *testing.T) {
		require.NoError(t, New[nvRootPtr]().Validate(&nvRootPtr{Leaf: nil}))
	})
	t.Run("value field, pointer receiver", func(t *testing.T) {
		err := New[nvRootVal]().Validate(&nvRootVal{Leaf: nvLeaf{Value: -1}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootVal]().Validate(&nvRootVal{Leaf: nvLeaf{Value: 1}}))
	})
	t.Run("value field, value receiver", func(t *testing.T) {
		err := New[nvRootByValue]().Validate(&nvRootByValue{Leaf: nvLeafByValue{Value: -1}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootByValue]().Validate(&nvRootByValue{Leaf: nvLeafByValue{Value: 1}}))
	})
	t.Run("two levels deep", func(t *testing.T) {
		err := New[nvRootDeep]().Validate(&nvRootDeep{Middle: &nvMiddle{Leaf: &nvLeaf{Value: -1}}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootDeep]().Validate(&nvRootDeep{Middle: &nvMiddle{Leaf: &nvLeaf{Value: 1}}}))
	})
	t.Run("slice elements", func(t *testing.T) {
		err := New[nvRootSlice]().Validate(&nvRootSlice{Leaves: []nvLeaf{{Value: 1}, {Value: -1}}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootSlice]().Validate(&nvRootSlice{Leaves: []nvLeaf{{Value: 1}, {Value: 2}}}))
	})
	t.Run("pointer slice elements", func(t *testing.T) {
		err := New[nvRootPtrSlice]().Validate(&nvRootPtrSlice{Leaves: []*nvLeaf{{Value: 1}, {Value: -1}}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootPtrSlice]().Validate(&nvRootPtrSlice{Leaves: []*nvLeaf{{Value: 1}}}))
	})
	t.Run("map values", func(t *testing.T) {
		err := New[nvRootMap]().Validate(&nvRootMap{Leaves: map[string]nvLeaf{"a": {Value: 1}, "b": {Value: -1}}})
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		require.NoError(t, New[nvRootMap]().Validate(&nvRootMap{Leaves: map[string]nvLeaf{"a": {Value: 1}}}))
	})
}

// TestValidatable_NestedUnmarshal runs the same shapes through Unmarshal (JSON in, Validate afterwards).
func TestValidatable_NestedUnmarshal(t *testing.T) {
	t.Run("pointer field, pointer receiver", func(t *testing.T) {
		_, err := New[nvRootPtr]().Unmarshal([]byte(`{"leaf":{"value":-1}}`))
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
		_, err = New[nvRootPtr]().Unmarshal([]byte(`{"leaf":{"value":1}}`))
		require.NoError(t, err)
	})
	t.Run("absent pointer field", func(t *testing.T) {
		_, err := New[nvRootPtr]().Unmarshal([]byte(`{}`))
		require.NoError(t, err)
	})
	t.Run("value field, value receiver", func(t *testing.T) {
		_, err := New[nvRootByValue]().Unmarshal([]byte(`{"leaf":{"value":-1}}`))
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
	})
	t.Run("two levels deep", func(t *testing.T) {
		_, err := New[nvRootDeep]().Unmarshal([]byte(`{"middle":{"leaf":{"value":-1}}}`))
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
	})
	t.Run("slice elements", func(t *testing.T) {
		_, err := New[nvRootSlice]().Unmarshal([]byte(`{"leaves":[{"value":1},{"value":-1}]}`))
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
	})
	t.Run("map values", func(t *testing.T) {
		_, err := New[nvRootMap]().Unmarshal([]byte(`{"leaves":{"a":{"value":1},"b":{"value":-1}}}`))
		require.Error(t, err)
		assert.Contains(t, nvMessages(t, err), nvNegativeMessage)
	})
}
