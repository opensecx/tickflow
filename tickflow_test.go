package tickflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTickFlow(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		tf, err := NewTickFlow("")
		assert.Nil(t, tf)
		assert.ErrorIs(t, err, ErrEmptyKey)
	})

	t.Run("valid key returns client", func(t *testing.T) {
		tf, err := NewTickFlow("test-api-key")
		assert.NotNil(t, tf)
		assert.NoError(t, err)
		assert.Equal(t, "test-api-key", tf.xApiKey)
		assert.Equal(t, defaultBaseURL, tf.baseURL)
	})
}
