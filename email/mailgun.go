package email

import (
	"context"
	"fmt"
	"gopkg.in/mailgun/mailgun-go.v3"
	"net/url"
	"time"
)

const (
	welcomeSubject = "Welcome to Tataruma.com"
	resetSubject   = "Instruction for resetting your password."

	resetBaseURL = "https://www.tataruma.com/reset"
)

const welcomeText = `Hi there!

Welcome to Tataruma.com! We really hope you enjoy using our application!

Best,
Dwi

`
const welcomeHTML = `Hi there!<br/>
<br/>
Welcome to <a href="https://www.tataruma.com">Tataruma.com</a>!<br/>

We really hope you enjoy using our application!</br>
<br/>
Best,<br/>
Dwi
`

const resetTextTmpl = `Hi there!
It appears that you requested password reset. If this was you, please follow the link below:

%s

If you are asked for a token, please use the following value:

%s

If you did not request a password reset you can safely ignore this email and your account will not be changed.

Best,
Tataruma Support
`

const resetHTMLTmpl = `Hi there!<br/>
<br/>
It appears that you requested password reset. If this was you, please follow the link below:
<br/>
<a href="%s">%s</a><br/>
<br/>
If you are asked for a token, please use the following value:<br/>

%s<br/>
<br/>
If you did not request a password reset you can safely ignore this email and your account will not be changed.<br/>
<br/>
Best,<br/>
Tataruma Support<br/>
`

type Client struct {
	from string
	mg   mailgun.Mailgun
}

type ClientConfig func(*Client)

func NewClient(opts ...ClientConfig) *Client {
	client := Client{
		from: "no-reply@tataruma.com",
	}

	for _, opt := range opts {
		opt(&client)
	}

	return &client
}

func WithMailgun(domain, apiKey string) ClientConfig {
	return func(client *Client) {
		mg := mailgun.NewMailgun(domain, apiKey)
		client.mg = mg
	}
}

func WithSender(toName, toEmail string) ClientConfig {
	return func(client *Client) {
		client.from = buildEmail(toName, toEmail)
	}
}

func (c *Client) Welcome(toName, toEmail string) error {
	message := c.mg.NewMessage(c.from, welcomeSubject, welcomeText, buildEmail(toName, toEmail))
	message.SetHtml(welcomeHTML)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := c.mg.Send(ctx, message)
	return err
}

func (c *Client) ResetPw(toEmail, token string) error {
	v := url.Values{}
	v.Set("token", token)
	resetUrl := resetBaseURL + "?" + v.Encode()

	resetText := fmt.Sprintf(resetTextTmpl, resetUrl, token)

	message := c.mg.NewMessage(c.from, resetSubject, resetText, toEmail)
	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetUrl, resetUrl, token)
	message.SetHtml(resetHTML)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := c.mg.Send(ctx, message)

	return err
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
