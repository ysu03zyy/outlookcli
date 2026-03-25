package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// MailFrom lists messages from a sender (uses Graph search).
func (c *Client) MailFrom(ctx context.Context, sender string, count int) ([]map[string]any, error) {
	if count < 1 {
		count = 20
	}
	v := url.Values{}
	v.Set("$search", fmt.Sprintf(`"from:%s"`, strings.ReplaceAll(sender, `"`, ``)))
	v.Set("$top", fmt.Sprintf("%d", count))
	v.Set("$select", "id,subject,from,receivedDateTime,isRead")
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
			IsRead           bool   `json:"isRead"`
		} `json:"value"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return nil, fmt.Errorf("graph: %s", resp.Error.Message)
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

// MailRead returns message details for printing.
func (c *Client) MailRead(ctx context.Context, suffix string) (map[string]any, error) {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return nil, err
	}
	u := graphBase + "/messages/" + url.PathEscape(id) + "?$select=subject,from,receivedDateTime,body,toRecipients"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var raw struct {
		Subject          string `json:"subject"`
		From             *struct{ EmailAddress *struct{ Name string `json:"name"`; Address string `json:"address"` } `json:"emailAddress"` } `json:"from"`
		ReceivedDateTime string `json:"receivedDateTime"`
		Body             *struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients []struct {
			EmailAddress *struct{ Address string `json:"address"` } `json:"emailAddress"`
		} `json:"toRecipients"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	body := ""
	if raw.Body != nil {
		if strings.EqualFold(raw.Body.ContentType, "html") {
			body = StripHTML(raw.Body.Content)
		} else {
			body = raw.Body.Content
		}
	}
	if len(body) > 2000 {
		body = body[:2000]
	}
	from := map[string]any{}
	if raw.From != nil && raw.From.EmailAddress != nil {
		from["name"] = raw.From.EmailAddress.Name
		from["address"] = raw.From.EmailAddress.Address
	}
	to := []string{}
	for _, r := range raw.ToRecipients {
		if r.EmailAddress != nil {
			to = append(to, r.EmailAddress.Address)
		}
	}
	return map[string]any{
		"subject": raw.Subject,
		"from":    from,
		"to":      to,
		"date":    raw.ReceivedDateTime,
		"body":    body,
	}, nil
}

// MailFolders lists mail folders.
func (c *Client) MailFolders(ctx context.Context) ([]map[string]any, error) {
	u := graphBase + "/mailFolders"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			DisplayName     string `json:"displayName"`
			TotalItemCount  int    `json:"totalItemCount"`
			UnreadItemCount int    `json:"unreadItemCount"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for _, f := range resp.Value {
		out = append(out, map[string]any{
			"name":   f.DisplayName,
			"total":  f.TotalItemCount,
			"unread": f.UnreadItemCount,
		})
	}
	return out, nil
}

// MailStats returns inbox folder stats.
func (c *Client) MailStats(ctx context.Context) (map[string]any, error) {
	u := graphBase + "/mailFolders/inbox"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var raw struct {
		DisplayName     string `json:"displayName"`
		TotalItemCount  int    `json:"totalItemCount"`
		UnreadItemCount int    `json:"unreadItemCount"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	return map[string]any{
		"folder": raw.DisplayName,
		"total":  raw.TotalItemCount,
		"unread": raw.UnreadItemCount,
	}, nil
}

// MailSend sends a new message.
func (c *Client) MailSend(ctx context.Context, to, subject, body string) error {
	payload := map[string]any{
		"message": map[string]any{
			"subject": subject,
			"body":    map[string]any{"contentType": "Text", "content": body},
			"toRecipients": []any{
				map[string]any{"emailAddress": map[string]any{"address": to}},
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	respBody, code, err := c.graphJSON(ctx, "POST", graphBase+"/sendMail", bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code != 202 {
		return fmt.Errorf("sendMail: %d: %s", code, string(respBody))
	}
	return nil
}

// MailReply sends a reply to a message.
func (c *Client) MailReply(ctx context.Context, suffix, comment string) error {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return err
	}
	payload := map[string]any{"comment": comment}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := graphBase + "/messages/" + url.PathEscape(id) + "/reply"
	respBody, code, err := c.graphJSON(ctx, "POST", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code != 202 {
		return fmt.Errorf("reply: %d: %s", code, string(respBody))
	}
	return nil
}

func (c *Client) mailPatch(ctx context.Context, suffix string, patch map[string]any) error {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return err
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	u := graphBase + "/messages/" + url.PathEscape(id)
	respBody, code, err := c.graphJSON(ctx, "PATCH", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("patch: %d: %s", code, string(respBody))
	}
	return nil
}

// MailMarkRead marks a message read.
func (c *Client) MailMarkRead(ctx context.Context, suffix string) error {
	return c.mailPatch(ctx, suffix, map[string]any{"isRead": true})
}

// MailMarkUnread marks a message unread.
func (c *Client) MailMarkUnread(ctx context.Context, suffix string) error {
	return c.mailPatch(ctx, suffix, map[string]any{"isRead": false})
}

// MailFlag flags a message.
func (c *Client) MailFlag(ctx context.Context, suffix string) error {
	return c.mailPatch(ctx, suffix, map[string]any{"flag": map[string]any{"flagStatus": "flagged"}})
}

// MailUnflag removes flag.
func (c *Client) MailUnflag(ctx context.Context, suffix string) error {
	return c.mailPatch(ctx, suffix, map[string]any{"flag": map[string]any{"flagStatus": "notFlagged"}})
}

// MailDelete moves message to deleted items.
func (c *Client) MailDelete(ctx context.Context, suffix string) error {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return err
	}
	payload := map[string]any{"destinationId": "deleteditems"}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := graphBase + "/messages/" + url.PathEscape(id) + "/move"
	respBody, code, err := c.graphJSON(ctx, "POST", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("move: %d: %s", code, string(respBody))
	}
	return nil
}

// MailArchive moves message to archive folder.
func (c *Client) MailArchive(ctx context.Context, suffix string) error {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return err
	}
	payload := map[string]any{"destinationId": "archive"}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := graphBase + "/messages/" + url.PathEscape(id) + "/move"
	respBody, code, err := c.graphJSON(ctx, "POST", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("move: %d: %s", code, string(respBody))
	}
	return nil
}

// MailMove moves a message to a folder by display name (case-insensitive).
func (c *Client) MailMove(ctx context.Context, suffix, folderName string) error {
	id, err := c.ResolveMessageID(ctx, suffix)
	if err != nil {
		return err
	}
	folders, err := c.mailFolderList(ctx)
	if err != nil {
		return err
	}
	folderLower := strings.ToLower(strings.TrimSpace(folderName))
	var folderID string
	for _, f := range folders {
		if strings.ToLower(f.Name) == folderLower {
			folderID = f.ID
			break
		}
	}
	if folderID == "" {
		names := make([]string, 0, len(folders))
		for _, f := range folders {
			names = append(names, f.Name)
		}
		return fmt.Errorf("folder not found: %q; available: %s", folderName, strings.Join(names, ", "))
	}
	payload := map[string]any{"destinationId": folderID}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := graphBase + "/messages/" + url.PathEscape(id) + "/move"
	respBody, code, err := c.graphJSON(ctx, "POST", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("move: %d: %s", code, string(respBody))
	}
	return nil
}

type folderRow struct {
	ID   string
	Name string
}

func (c *Client) mailFolderList(ctx context.Context) ([]folderRow, error) {
	u := graphBase + "/mailFolders"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]folderRow, 0, len(resp.Value))
	for _, f := range resp.Value {
		out = append(out, folderRow{ID: f.ID, Name: f.DisplayName})
	}
	return out, nil
}
