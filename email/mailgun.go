package email

import (
	"context"
	"fmt"
	"gopkg.in/mailgun/mailgun-go.v3"
	"time"
)

const (
	welcomeSubject = "Welcome to Tataruma.com"
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

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
