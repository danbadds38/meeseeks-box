package formatter

import (
	"strings"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/template"
)

// Default colors
const (
	DefaultInfoColorMessage    = ""
	DefaultSuccessColorMessage = "good"
	DefaultWarningColorMessage = "warning"
	DefaultErrColorMessage     = "danger"
)

// MessageColors contains the configured reply message colora
type MessageColors struct {
	Info    string `yaml:"info"`
	Success string `yaml:"success"`
	Error   string `yaml:"error"`
}

// FormatConfig contains the formatting configurations
type FormatConfig struct {
	Colors     MessageColors     `yaml:"colors"`
	ReplyStyle map[string]string `yaml:"reply_styles"`
}

// Formatter keeps the colors and templates used to format a reply message
type Formatter struct {
	colors     MessageColors
	templates  *template.TemplatesBuilder
	replyStyle replyStyle
}

var formatter *Formatter

// Configure sets up the singleton formatter
func Configure(messages map[string][]string, cnf FormatConfig) {
	builder := template.NewBuilder().WithMessages(messages)
	formatter = &Formatter{
		replyStyle: replyStyle{cnf.ReplyStyle},
		colors:     cnf.Colors,
		templates:  builder,
	}
}

// Get returns the configured singleton formatter
func Get() *Formatter {
	return formatter
}

// Templates returns a clone of the default templates ready to be consumed
func (f Formatter) Templates() template.Templates {
	return f.templates.Clone().Build()
}

// WithTemplates returns a clone of the default templates with the templates
// passed as argument applied on top
func (f Formatter) WithTemplates(templates map[string]string) template.Templates {
	return f.templates.Clone().WithTemplates(templates).Build()
}

// HandshakeReply creates a reply for a handshake message
func (f Formatter) HandshakeReply(req meeseeks.Request) Reply {
	return f.newReplier(template.Handshake, req)
}

// UnknownCommandReply creates a reply for an UnknownCommand error message
func (f Formatter) UnknownCommandReply(req meeseeks.Request) Reply {
	return f.newReplier(template.UnknownCommand, req)
}

// UnauthorizedCommandReply creates a reply for an unauthorized command error message
func (f Formatter) UnauthorizedCommandReply(req meeseeks.Request) Reply {
	return f.newReplier(template.Unauthorized, req)
}

// FailureReply creates a reply for a generic command error message
func (f Formatter) FailureReply(req meeseeks.Request, err error) Reply {
	return f.newReplier(template.Failure, req).WithError(err)
}

// SuccessReply creates a reply for a generic command success message
func (f Formatter) SuccessReply(req meeseeks.Request) Reply {
	return f.newReplier(template.Success, req)
}

func (f Formatter) newReplier(action string, req meeseeks.Request) Reply {
	return Reply{
		action:  action,
		request: req,

		templates: f.templates.Clone(),
		style:     f.replyStyle.Get(action),
		colors:    f.colors,
	}
}

type replyStyle struct {
	styles map[string]string
}

func (r replyStyle) Get(mode string) string {
	switch mode {
	case template.Handshake,
		template.UnknownCommand,
		template.Unauthorized,
		template.Failure,
		template.Success:

		if style, ok := r.styles[mode]; ok {
			return style
		}
	}
	return ""
}

// Reply represents all the data necessary to send a reply message
type Reply struct {
	action  string
	request meeseeks.Request
	output  string
	err     error

	colors    MessageColors
	templates *template.TemplatesBuilder
	style     string
}

// WithOutput stores the text payload to render in the reply
func (r Reply) WithOutput(output string) Reply {
	r.output = output
	return r
}

// WithError stores an error to render
func (r Reply) WithError(err error) Reply {
	r.err = err
	return r
}

// Render renders the message returning the rendered text, or an error if something goes wrong.
func (r Reply) Render() (string, error) {
	payload := make(map[string]interface{})
	payload["command"] = r.request.Command
	payload["args"] = strings.Join(r.request.Args, " ")

	payload["user"] = r.request.Username
	payload["userlink"] = r.request.UserLink
	payload["userid"] = r.request.UserID
	payload["channel"] = r.request.Channel
	payload["channellink"] = r.request.ChannelLink
	payload["channelid"] = r.request.ChannelID
	payload["isim"] = r.request.IsIM

	payload["error"] = r.err
	payload["output"] = r.output

	return r.templates.Build().Render(r.action, payload)
}

// ChannelID returns the channel ID in which to reply
func (r Reply) ChannelID() string {
	return r.request.ChannelID
}

// ReplyStyle returns the style to use to reply
func (r Reply) ReplyStyle() string {
	return r.style
}

// Color returns the color to use when decorating the reply
func (r Reply) Color() string {
	switch r.action {
	case template.Handshake:
		return r.colors.Info
	case template.UnknownCommand, template.Unauthorized, template.Failure:
		return r.colors.Error
	default:
		return r.colors.Success
	}
}
