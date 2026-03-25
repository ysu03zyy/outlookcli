package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ResolveEventID finds full event id by suffix (recent events).
func (c *Client) ResolveEventID(ctx context.Context, suffix string) (string, error) {
	suffix = strings.TrimSpace(suffix)
	if suffix == "" {
		return "", fmt.Errorf("empty event id")
	}
	u := graphBase + "/calendar/events?$top=50&$select=id"
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
	for _, e := range resp.Value {
		if strings.HasSuffix(e.ID, suffix) {
			return e.ID, nil
		}
	}
	return "", fmt.Errorf("event not found for id suffix %q", suffix)
}

func (c *Client) calTimeZone() string {
	if c.TZName != "" {
		return c.TZName
	}
	return "UTC"
}

// CalendarEvents lists upcoming events (newest first).
func (c *Client) CalendarEvents(ctx context.Context, count int) ([]map[string]any, error) {
	if count < 1 {
		count = 10
	}
	u := fmt.Sprintf("%s/calendar/events?$top=%d&$orderby=start/dateTime%%20desc&$select=id,subject,start,end,location,isAllDay", graphBase, count)
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	return c.parseEventList(b)
}

// CalendarToday lists today's events via calendarView.
func (c *Client) CalendarToday(ctx context.Context) ([]map[string]any, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24*time.Hour - time.Nanosecond)
	return c.calendarView(ctx, start, end)
}

// CalendarWeek lists events in the next 7 days from start of today UTC.
func (c *Client) CalendarWeek(ctx context.Context) ([]map[string]any, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)
	return c.calendarView(ctx, start, end)
}

func (c *Client) calendarView(ctx context.Context, start, end time.Time) ([]map[string]any, error) {
	v := url.Values{}
	v.Set("startDateTime", start.Format(time.RFC3339))
	v.Set("endDateTime", end.Format(time.RFC3339))
	v.Set("$orderby", "start/dateTime")
	v.Set("$select", "id,subject,start,end,location")
	u := graphBase + "/calendarView?" + v.Encode()
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	return c.parseEventList(b)
}

func (c *Client) parseEventList(b []byte) ([]map[string]any, error) {
	var resp struct {
		Value []struct {
			ID       string `json:"id"`
			Subject  string `json:"subject"`
			Start    *struct{ DateTime string `json:"dateTime"` } `json:"start"`
			End      *struct{ DateTime string `json:"dateTime"` } `json:"end"`
			Location *struct{ DisplayName string `json:"displayName"` } `json:"location"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for i, e := range resp.Value {
		st, en := "", ""
		if e.Start != nil {
			st = e.Start.DateTime
			if len(st) > 16 {
				st = st[:16]
			}
		}
		if e.End != nil {
			en = e.End.DateTime
			if len(en) > 16 {
				en = en[:16]
			}
		}
		loc := ""
		if e.Location != nil {
			loc = e.Location.DisplayName
		}
		out = append(out, map[string]any{
			"n":        i + 1,
			"subject":  e.Subject,
			"start":    st,
			"end":      en,
			"location": loc,
			"id":       ShortID(e.ID),
		})
	}
	return out, nil
}

// CalendarRead returns event details.
func (c *Client) CalendarRead(ctx context.Context, suffix string) (map[string]any, error) {
	id, err := c.ResolveEventID(ctx, suffix)
	if err != nil {
		return nil, err
	}
	u := graphBase + "/calendar/events/" + url.PathEscape(id)
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var raw struct {
		Subject   string `json:"subject"`
		Start     *struct{ DateTime string `json:"dateTime"` } `json:"start"`
		End       *struct{ DateTime string `json:"dateTime"` } `json:"end"`
		Location  *struct{ DisplayName string `json:"displayName"` } `json:"location"`
		Body *struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		Attendees []struct {
			EmailAddress *struct{ Address string `json:"address"` } `json:"emailAddress"`
		} `json:"attendees"`
		IsOnlineMeeting bool   `json:"isOnlineMeeting"`
		OnlineMeeting *struct {
			JoinURL string `json:"joinUrl"`
		} `json:"onlineMeeting"`
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
	if len(body) > 500 {
		body = body[:500]
	}
	att := []string{}
	for _, a := range raw.Attendees {
		if a.EmailAddress != nil {
			att = append(att, a.EmailAddress.Address)
		}
	}
	st, en := "", ""
	if raw.Start != nil {
		st = raw.Start.DateTime
	}
	if raw.End != nil {
		en = raw.End.DateTime
	}
	loc := ""
	if raw.Location != nil {
		loc = raw.Location.DisplayName
	}
	link := ""
	if raw.OnlineMeeting != nil {
		link = raw.OnlineMeeting.JoinURL
	}
	return map[string]any{
		"subject":   raw.Subject,
		"start":     st,
		"end":       en,
		"location":  loc,
		"body":      body,
		"attendees": att,
		"isOnline":  raw.IsOnlineMeeting,
		"link":      link,
	}, nil
}

// CalendarCreate creates an event (local datetime strings, no Z).
func (c *Client) CalendarCreate(ctx context.Context, subject, start, end, location string) (map[string]any, error) {
	tz := c.calTimeZone()
	payload := map[string]any{
		"subject": subject,
		"start":   map[string]any{"dateTime": start, "timeZone": tz},
		"end":     map[string]any{"dateTime": end, "timeZone": tz},
	}
	if strings.TrimSpace(location) != "" {
		payload["location"] = map[string]any{"displayName": location}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	respBody, code, err := c.graphJSON(ctx, "POST", graphBase+"/calendar/events", bytes.NewReader(b), "application/json")
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("create: %d: %s", code, string(respBody))
	}
	var ev struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
		Start   *struct{ DateTime string `json:"dateTime"` } `json:"start"`
		End     *struct{ DateTime string `json:"dateTime"` } `json:"end"`
	}
	if err := json.Unmarshal(respBody, &ev); err != nil {
		return nil, err
	}
	st, en := "", ""
	if ev.Start != nil {
		st = ev.Start.DateTime
		if len(st) > 16 {
			st = st[:16]
		}
	}
	if ev.End != nil {
		en = ev.End.DateTime
		if len(en) > 16 {
			en = en[:16]
		}
	}
	return map[string]any{
		"status":  "event created",
		"subject": ev.Subject,
		"start":   st,
		"end":     en,
		"id":      ShortID(ev.ID),
	}, nil
}

// CalendarQuick creates a 1-hour event; if start is empty, uses now in local TZ formatting.
func (c *Client) CalendarQuick(ctx context.Context, subject, start string) (map[string]any, error) {
	tz := c.calTimeZone()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	var st time.Time
	if strings.TrimSpace(start) == "" {
		st = time.Now().In(loc)
	} else {
		st, err = time.ParseInLocation("2006-01-02T15:04", start, loc)
		if err != nil {
			return nil, fmt.Errorf("parse start time: %w (use YYYY-MM-DDTHH:MM)", err)
		}
	}
	en := st.Add(time.Hour)
	fs := st.Format("2006-01-02T15:04")
	fe := en.Format("2006-01-02T15:04")
	return c.CalendarCreate(ctx, subject, fs, fe, "")
}

// CalendarDelete deletes an event.
func (c *Client) CalendarDelete(ctx context.Context, suffix string) error {
	id, err := c.ResolveEventID(ctx, suffix)
	if err != nil {
		return err
	}
	u := graphBase + "/calendar/events/" + url.PathEscape(id)
	_, code, err := c.graphJSON(ctx, "DELETE", u, nil, "")
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete: %d", code)
	}
	return nil
}

// CalendarUpdate patches subject, location, start, or end.
func (c *Client) CalendarUpdate(ctx context.Context, suffix, field, value string) (map[string]any, error) {
	id, err := c.ResolveEventID(ctx, suffix)
	if err != nil {
		return nil, err
	}
	tz := c.calTimeZone()
	var patch map[string]any
	switch strings.ToLower(field) {
	case "subject":
		patch = map[string]any{"subject": value}
	case "location":
		patch = map[string]any{"location": map[string]any{"displayName": value}}
	case "start":
		patch = map[string]any{"start": map[string]any{"dateTime": value, "timeZone": tz}}
	case "end":
		patch = map[string]any{"end": map[string]any{"dateTime": value, "timeZone": tz}}
	default:
		return nil, fmt.Errorf("unknown field %q (use subject, location, start, end)", field)
	}
	b, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	u := graphBase + "/calendar/events/" + url.PathEscape(id)
	respBody, code, err := c.graphJSON(ctx, "PATCH", u, bytes.NewReader(b), "application/json")
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("update: %d: %s", code, string(respBody))
	}
	var ev struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
		Start   *struct{ DateTime string `json:"dateTime"` } `json:"start"`
		End     *struct{ DateTime string `json:"dateTime"` } `json:"end"`
	}
	if err := json.Unmarshal(respBody, &ev); err != nil {
		return nil, err
	}
	st, en := "", ""
	if ev.Start != nil {
		st = ev.Start.DateTime
		if len(st) > 16 {
			st = st[:16]
		}
	}
	if ev.End != nil {
		en = ev.End.DateTime
		if len(en) > 16 {
			en = en[:16]
		}
	}
	return map[string]any{
		"status":  "event updated",
		"subject": ev.Subject,
		"start":   st,
		"end":     en,
		"id":      ShortID(ev.ID),
	}, nil
}

// CalendarList lists calendar containers.
func (c *Client) CalendarList(ctx context.Context) ([]map[string]any, error) {
	u := graphBase + "/calendars"
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			Name    string `json:"name"`
			Color   string `json:"color"`
			CanEdit bool   `json:"canEdit"`
			ID      string `json:"id"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(resp.Value))
	for _, cal := range resp.Value {
		out = append(out, map[string]any{
			"name":    cal.Name,
			"color":   cal.Color,
			"canEdit": cal.CanEdit,
			"id":      ShortID(cal.ID),
		})
	}
	return out, nil
}

// CalendarFree checks busy/free in a range (start/end as YYYY-MM-DDTHH:MM, interpreted with calendar TZ).
func (c *Client) CalendarFree(ctx context.Context, start, end string) (map[string]any, error) {
	tz := c.calTimeZone()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	st, err := time.ParseInLocation("2006-01-02T15:04", start, loc)
	if err != nil {
		return nil, fmt.Errorf("parse start: %w", err)
	}
	en, err := time.ParseInLocation("2006-01-02T15:04", end, loc)
	if err != nil {
		return nil, fmt.Errorf("parse end: %w", err)
	}
	v := url.Values{}
	v.Set("startDateTime", st.UTC().Format(time.RFC3339))
	v.Set("endDateTime", en.UTC().Format(time.RFC3339))
	v.Set("$select", "subject,start,end")
	u := graphBase + "/calendarView?" + v.Encode()
	b, code, err := c.graphGET(ctx, u)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("graph: %d: %s", code, string(b))
	}
	var resp struct {
		Value []struct {
			Subject string `json:"subject"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}
	if len(resp.Value) == 0 {
		return map[string]any{"status": "free", "start": start, "end": end}, nil
	}
	names := make([]string, 0, len(resp.Value))
	for _, e := range resp.Value {
		names = append(names, e.Subject)
	}
	return map[string]any{"status": "busy", "events": names}, nil
}
