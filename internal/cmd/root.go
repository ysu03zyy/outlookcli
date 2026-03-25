package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/ysu03zyy/outlookcli/internal/config"
	"github.com/ysu03zyy/outlookcli/internal/graph"
)

// Version is set by -ldflags at build time.
var Version = "dev"

// RootFlags are global flags.
type RootFlags struct {
	ConfigDir string `name:"config-dir" env:"OUTLOOK_CONFIG_DIR" help:"Config directory (default: ~/.outlook-mcp)"`
	JSON      bool   `short:"j" help:"Print JSON to stdout"`
	Timezone  string `name:"timezone" short:"z" env:"OUTLOOK_TIMEZONE" help:"IANA timezone for calendar (default UTC)" default:"UTC"`
}

// CLI is the root command tree.
type CLI struct {
	RootFlags `embed:""`

	Version kong.VersionFlag `short:"V" help:"Print version"`

	Token struct {
		Refresh tokenRefreshCmd `cmd:"" name:"refresh" help:"Refresh OAuth access token"`
		Test    tokenTestCmd    `cmd:"" name:"test" help:"Test Microsoft Graph connection"`
		Get     tokenGetCmd     `cmd:"" name:"get" help:"Print current access token (after refresh)"`
	} `cmd:"" name:"token" help:"OAuth token helpers"`

	Mail struct {
		Inbox     mailInboxCmd   `cmd:"" name:"inbox" help:"List latest inbox messages"`
		Unread    mailUnreadCmd  `cmd:"" name:"unread" help:"List unread messages"`
		Search    mailSearchCmd  `cmd:"" name:"search" help:"Search messages"`
		From      mailFromCmd    `cmd:"" name:"from" help:"List messages from sender"`
		Read      mailReadCmd    `cmd:"" name:"read" help:"Read message body"`
		Folders   mailFoldersCmd `cmd:"" name:"folders" help:"List mail folders"`
		Stats     mailStatsCmd   `cmd:"" name:"stats" help:"Inbox statistics"`
		Send      mailSendCmd    `cmd:"" name:"send" help:"Send an email"`
		Reply     mailReplyCmd   `cmd:"" name:"reply" help:"Reply to a message"`
		MarkRead   mailMarkReadCmd   `cmd:"" name:"mark-read" help:"Mark message as read"`
		MarkUnread mailMarkUnreadCmd `cmd:"" name:"mark-unread" help:"Mark message as unread"`
		Flag       mailFlagCmd       `cmd:"" name:"flag" help:"Flag message"`
		Unflag     mailUnflagCmd     `cmd:"" name:"unflag" help:"Remove flag"`
		Delete     mailDeleteCmd     `cmd:"" name:"delete" help:"Move message to trash"`
		Archive    mailArchiveCmd    `cmd:"" name:"archive" help:"Archive message"`
		Move       mailMoveCmd       `cmd:"" name:"move" help:"Move message to folder"`
	} `cmd:"" name:"mail" help:"Outlook mail (Microsoft Graph)"`

	Calendar struct {
		Events    calEventsCmd  `cmd:"" name:"events" help:"List upcoming events"`
		Today     calTodayCmd   `cmd:"" name:"today" help:"Today's events"`
		Week      calWeekCmd    `cmd:"" name:"week" help:"This week's events"`
		Read      calReadCmd    `cmd:"" name:"read" help:"Show event details"`
		Create    calCreateCmd  `cmd:"" name:"create" help:"Create event (start/end: YYYY-MM-DDTHH:MM)"`
		Quick     calQuickCmd   `cmd:"" name:"quick" help:"Quick 1-hour event"`
		Delete    calDeleteCmd  `cmd:"" name:"delete" help:"Delete event"`
		Update    calUpdateCmd  `cmd:"" name:"update" help:"Update event field"`
		Calendars calCalsCmd    `cmd:"" name:"calendars" help:"List calendars"`
		Free      calFreeCmd    `cmd:"" name:"free" help:"Check free/busy in a time range"`
	} `cmd:"" name:"calendar" aliases:"cal" help:"Outlook calendar"`
}

type tokenRefreshCmd struct{}

func (tokenRefreshCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	app, err := config.LoadAppConfig(dir)
	if err != nil {
		return err
	}
	if _, err := graph.Refresh(ctx, dir, app); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Token refreshed successfully")
	return nil
}

type tokenTestCmd struct{}

func (tokenTestCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	c := &graph.Client{Dir: dir, TZName: root.Timezone}
	total, unread, err := c.TestInbox(ctx)
	if err != nil {
		return err
	}
	if root.JSON {
		return emitJSON(map[string]any{"total": total, "unread": unread, "ok": true})
	}
	fmt.Fprintf(os.Stdout, "Connected. Inbox: %d messages (%d unread)\n", total, unread)
	return nil
}

type tokenGetCmd struct{}

func (tokenGetCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	tok, err := graph.EnsureAccessToken(ctx, dir)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, tok)
	return nil
}

type mailInboxCmd struct {
	Count int `short:"n" default:"10" help:"Number of messages"`
}

func (c *mailInboxCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.MailInbox(ctx, c.Count)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type mailUnreadCmd struct {
	Count int `short:"n" default:"20" help:"Number of messages"`
}

func (c *mailUnreadCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.MailUnread(ctx, c.Count)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type mailSearchCmd struct {
	Query string `arg:"" help:"Search query"`
	Count int    `short:"n" default:"20" help:"Max results"`
}

func (c *mailSearchCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.MailSearch(ctx, c.Query, c.Count)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type mailFromCmd struct {
	Sender string `arg:"" help:"Sender email"`
	Count  int    `short:"n" default:"20" help:"Max results"`
}

func (c *mailFromCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.MailFrom(ctx, c.Sender, c.Count)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type mailReadCmd struct {
	ID string `arg:"" help:"Message id (suffix from list output)"`
}

func (c *mailReadCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.MailRead(ctx, c.ID)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type mailFoldersCmd struct{}

func (mailFoldersCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.MailFolders(ctx)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type mailStatsCmd struct{}

func (mailStatsCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.MailStats(ctx)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type mailSendCmd struct {
	To      string `required:"" short:"t" help:"Recipient email"`
	Subject string `required:"" short:"s" help:"Subject"`
	Body    string `required:"" short:"b" help:"Body (plain text)"`
}

func (c *mailSendCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailSend(ctx, c.To, c.Subject, c.Body); err != nil {
		return err
	}
	if root.JSON {
		return emitJSON(map[string]any{"status": "sent", "to": c.To, "subject": c.Subject})
	}
	fmt.Fprintf(os.Stdout, "Sent to %s: %s\n", c.To, c.Subject)
	return nil
}

type mailReplyCmd struct {
	ID   string `arg:"" help:"Message id suffix"`
	Body string `arg:"" help:"Reply body"`
}

func (c *mailReplyCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailReply(ctx, c.ID, c.Body); err != nil {
		return err
	}
	if root.JSON {
		return emitJSON(map[string]any{"status": "reply sent", "id": c.ID})
	}
	fmt.Fprintln(os.Stdout, "Reply sent")
	return nil
}

type mailMarkReadCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailMarkReadCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailMarkRead(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "marked as read", c.ID)
}

type mailMarkUnreadCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailMarkUnreadCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailMarkUnread(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "marked as unread", c.ID)
}

type mailFlagCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailFlagCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailFlag(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "flagged", c.ID)
}

type mailUnflagCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailUnflagCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailUnflag(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "unflagged", c.ID)
}

type mailDeleteCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailDeleteCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailDelete(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "moved to trash", c.ID)
}

type mailArchiveCmd struct {
	ID string `arg:"" help:"Message id suffix"`
}

func (c *mailArchiveCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailArchive(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "archived", c.ID)
}

type mailMoveCmd struct {
	ID     string `arg:"" help:"Message id suffix"`
	Folder string `arg:"" help:"Folder display name"`
}

func (c *mailMoveCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.MailMove(ctx, c.ID, c.Folder); err != nil {
		return err
	}
	if root.JSON {
		return emitJSON(map[string]any{"status": "moved", "folder": c.Folder, "id": c.ID})
	}
	fmt.Fprintf(os.Stdout, "Moved to %s\n", c.Folder)
	return nil
}

type calEventsCmd struct {
	Count int `short:"n" default:"10" help:"Number of events"`
}

func (c *calEventsCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.CalendarEvents(ctx, c.Count)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type calTodayCmd struct{}

func (calTodayCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.CalendarToday(ctx)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type calWeekCmd struct{}

func (calWeekCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.CalendarWeek(ctx)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type calReadCmd struct {
	ID string `arg:"" help:"Event id suffix"`
}

func (c *calReadCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.CalendarRead(ctx, c.ID)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type calCreateCmd struct {
	Subject  string `arg:"" help:"Event title"`
	Start    string `arg:"" help:"Start YYYY-MM-DDTHH:MM"`
	End      string `arg:"" help:"End YYYY-MM-DDTHH:MM"`
	Location string `name:"location" short:"l" help:"Location (optional)"`
}

func (c *calCreateCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.CalendarCreate(ctx, c.Subject, c.Start, c.End, c.Location)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type calQuickCmd struct {
	Subject string `arg:"" help:"Event title"`
	Start   string `name:"start" short:"s" help:"Start YYYY-MM-DDTHH:MM (default: now)"`
}

func (c *calQuickCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.CalendarQuick(ctx, c.Subject, c.Start)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type calDeleteCmd struct {
	ID string `arg:"" help:"Event id suffix"`
}

func (c *calDeleteCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	if err := client.CalendarDelete(ctx, c.ID); err != nil {
		return err
	}
	return emitOK(root, "event deleted", c.ID)
}

type calUpdateCmd struct {
	ID    string `arg:"" help:"Event id suffix"`
	Field string `arg:"" help:"subject|location|start|end"`
	Value string `arg:"" help:"New value"`
}

func (c *calUpdateCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.CalendarUpdate(ctx, c.ID, c.Field, c.Value)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

type calCalsCmd struct{}

func (calCalsCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	rows, err := client.CalendarList(ctx)
	if err != nil {
		return err
	}
	return emitRows(root, rows)
}

type calFreeCmd struct {
	Start string `arg:"" help:"Start YYYY-MM-DDTHH:MM"`
	End   string `arg:"" help:"End YYYY-MM-DDTHH:MM"`
}

func (c *calFreeCmd) Run(ctx context.Context, root *RootFlags) error {
	dir, err := resolveDir(root)
	if err != nil {
		return err
	}
	client := &graph.Client{Dir: dir, TZName: root.Timezone}
	m, err := client.CalendarFree(ctx, c.Start, c.End)
	if err != nil {
		return err
	}
	return emitObject(root, m)
}

func resolveDir(root *RootFlags) (string, error) {
	if root.ConfigDir != "" {
		return filepath.Clean(root.ConfigDir), nil
	}
	return config.Dir()
}

func emitJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func emitRows(root *RootFlags, rows []map[string]any) error {
	if root.JSON {
		return emitJSON(rows)
	}
	for _, row := range rows {
		b, err := json.Marshal(row)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, string(b))
	}
	return nil
}

func emitObject(root *RootFlags, m map[string]any) error {
	if root.JSON {
		return emitJSON(m)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, string(b))
	return nil
}

func emitOK(root *RootFlags, status, id string) error {
	if root.JSON {
		return emitJSON(map[string]any{"status": status, "id": id})
	}
	fmt.Fprintf(os.Stdout, "%s (%s)\n", status, id)
	return nil
}

// Execute runs the CLI (args should not include the program name).
func Execute(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("outlookcli"),
		kong.Description("Outlook mail and calendar via Microsoft Graph"),
		kong.Vars(kong.Vars{"version": Version}),
	)
	if err != nil {
		return err
	}
	kctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	ctx := context.Background()
	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)
	return kctx.Run()
}
