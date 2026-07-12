package toolsets

import (
	"testing"

	"github.com/harvester/harvester-mcp-server/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllToolSets(t *testing.T) {
	c, err := client.NewClient(true, "")
	require.NoError(t, err)
	toolsets := allToolSets(c, false)

	assert.NotNil(t, toolsets)
	assert.Len(t, toolsets, 1, "should have exactly 1 toolset")
}
