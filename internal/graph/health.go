package graph

import (
	"context"
	"encoding/json"
	"fmt"
)

// TestInbox checks Graph connectivity using the inbox folder.
func (c *Client) TestInbox(ctx context.Context) (total, unread int, err error) {
	b, code, err := c.graphGET(ctx, graphBase+"/mailFolders/inbox")
	if err != nil {
		return 0, 0, err
	}
	if code < 200 || code >= 300 {
		return 0, 0, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var r struct {
		TotalItemCount  int `json:"totalItemCount"`
		UnreadItemCount int `json:"unreadItemCount"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return 0, 0, err
	}
	return r.TotalItemCount, r.UnreadItemCount, nil
}
