package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const graphBase = "https://graph.microsoft.com/v1.0/me"

// ResolveMessageID finds full message id by suffix match (recent messages).
func (c *Client) ResolveMessageID(ctx context.Context, suffix string) (string, error) {
	suffix = strings.TrimSpace(suffix)
	if suffix == "" {
		return "", fmt.Errorf("empty message id")
	}
	u := graphBase + "/messages?$top=100&$select=id"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return "", err
	}
	if code < 200 || code >= 300 {
		return "", fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", err
	}
	for _, m := range resp.Value {
		if strings.HasSuffix(m.ID, suffix) {
			return m.ID, nil
		}
	}
	return "", fmt.Errorf("message not found for id suffix %q (searched last 100 messages)", suffix)
}

// MailInbox lists latest inbox messages (JSON lines compatible with skill output shape).
func (c *Client) MailInbox(ctx context.Context, count int) ([]map[string]any, error) {
	if count < 1 {
		count = 10
	}
	u := fmt.Sprintf("%s/messages?$top=%d&$orderby=receivedDateTime%%20desc&$select=id,subject,from,receivedDateTime,isRead", graphBase, count)
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			ID                string `json:"id"`
			Subject           string `json:"subject"`
			From              *struct{ EmailAddress *struct{ Address string `json:"address"` } `json:"emailAddress"` } `json:"from"`
			ReceivedDateTime  string `json:"receivedDateTime"`
			IsRead            bool   `json:"isRead"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for i, m := range resp.Value {
		from := ""
		if m.From != nil && m.From.EmailAddress != nil {
			from = m.From.EmailAddress.Address
		}
		d := m.ReceivedDateTime
		if len(d) > 16 {
			d = d[:16]
		}
		out = append(out, map[string]any{
			"n":       i + 1,
			"subject": m.Subject,
			"from":    from,
			"date":    d,
			"read":    m.IsRead,
			"id":      ShortID(m.ID),
		})
	}
	return out, nil
}

// MailUnread lists unread messages.
func (c *Client) MailUnread(ctx context.Context, count int) ([]map[string]any, error) {
	if count < 1 {
		count = 20
	}
	u := fmt.Sprintf("%s/messages?$filter=isRead%%20eq%%20false&$top=%d&$orderby=receivedDateTime%%20desc&$select=id,subject,from,receivedDateTime", graphBase, count)
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			ID               string `json:"id"`
			Subject          string `json:"subject"`
			From             *struct{ EmailAddress *struct{ Address string `json:"address"` } } `json:"from"`
			ReceivedDateTime string `json:"receivedDateTime"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for i, m := range resp.Value {
		from := ""
		if m.From != nil && m.From.EmailAddress != nil {
			from = m.From.EmailAddress.Address
		}
		d := m.ReceivedDateTime
		if len(d) > 16 {
			d = d[:16]
		}
		out = append(out, map[string]any{
			"n":       i + 1,
			"subject": m.Subject,
			"from":    from,
			"date":    d,
			"id":      ShortID(m.ID),
		})
	}
	return out, nil
}

// MailSearch searches messages.
func (c *Client) MailSearch(ctx context.Context, query string, count int) ([]map[string]any, error) {
	if count < 1 {
		count = 20
	}
	v := url.Values{}
	v.Set("$search", fmt.Sprintf(`"%s"`, strings.ReplaceAll(query, `"`, ``)))
	v.Set("$top", fmt.Sprintf("%d", count))
	v.Set("$select", "id,subject,from,receivedDateTime")
	u := graphBase + "/messages?" + v.Encode()
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			ID               string `json:"id"`
			Subject          string `json:"subject"`
			From             *struct{ EmailAddress *struct{ Address string `json:"address"` } } `json:"from"`
			ReceivedDateTime string `json:"receivedDateTime"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for i, m := range resp.Value {
		from := ""
		if m.From != nil && m.From.EmailAddress != nil {
			from = m.From.EmailAddress.Address
		}
		d := m.ReceivedDateTime
		if len(d) > 16 {
			d = d[:16]
		}
		out = append(out, map[string]any{
			"n":       i + 1,
			"subject": m.Subject,
			"from":    from,
			"date":    d,
			"id":      ShortID(m.ID),
		})
	}
	return out, nil
}
