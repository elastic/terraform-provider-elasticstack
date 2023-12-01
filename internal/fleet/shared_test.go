package fleet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SortInputs(t *testing.T) {
	t.Run("WithExisting", func(t *testing.T) {
		existing := []any{
			map[string]any{"input_id": "A", "enabled": true},
			map[string]any{"input_id": "B", "enabled": true},
			map[string]any{"input_id": "C", "enabled": true},
			map[string]any{"input_id": "D", "enabled": true},
			map[string]any{"input_id": "E", "enabled": true},
		}

		incoming := []any{
			map[string]any{"input_id": "G", "enabled": true},
			map[string]any{"input_id": "F", "enabled": true},
			map[string]any{"input_id": "B", "enabled": true},
			map[string]any{"input_id": "E", "enabled": true},
			map[string]any{"input_id": "C", "enabled": true},
		}

		want := []any{
			map[string]any{"input_id": "B", "enabled": true},
			map[string]any{"input_id": "C", "enabled": true},
			map[string]any{"input_id": "E", "enabled": true},
			map[string]any{"input_id": "G", "enabled": true},
			map[string]any{"input_id": "F", "enabled": true},
		}

		sortInputs(incoming, existing)

		require.Equal(t, want, incoming)
	})

	t.Run("WithEmpty", func(t *testing.T) {
		var existing []any

		incoming := []any{
			map[string]any{"input_id": "G", "enabled": true},
			map[string]any{"input_id": "F", "enabled": true},
			map[string]any{"input_id": "B", "enabled": true},
			map[string]any{"input_id": "E", "enabled": true},
			map[string]any{"input_id": "C", "enabled": true},
		}

		want := []any{
			map[string]any{"input_id": "G", "enabled": true},
			map[string]any{"input_id": "F", "enabled": true},
			map[string]any{"input_id": "B", "enabled": true},
			map[string]any{"input_id": "E", "enabled": true},
			map[string]any{"input_id": "C", "enabled": true},
		}

		sortInputs(incoming, existing)

		require.Equal(t, want, incoming)
	})
}
