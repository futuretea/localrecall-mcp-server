package toolset

import (
	"github.com/futuretea/localrecall-mcp-server/pkg/client"
)

// LocalRecallClient wraps the LocalRecall API client for use in toolset
type LocalRecallClient struct {
	Client *client.Client
}
