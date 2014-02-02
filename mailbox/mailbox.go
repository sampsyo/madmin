// A simple IMAP email client that abstracts away the details of the IMAP
// protocol.
package mailbox

import (
	"bytes"
	"code.google.com/p/go-imap/go1/imap"
	"errors"
	"net/mail"
	"time"
)

// A connection to an IMAP server.
type Connection struct {
	client *imap.Client
}

// The metadata for an email message. Includes "envelope" information but not
// the body of the message.
type MessageMeta struct {
	// The message's unique identifier.
	UID uint32

	// The subject line.
	Subject string

	// The message's total size in bytes.
	Size uint32
}

func (conn *Connection) connect(host string) error {
	var err error
	conn.client, err = imap.DialTLS("imap.gmail.com", nil)
	return err
}

func (conn *Connection) login(user string, pass string) error {
	if conn.client.State() != imap.Login {
		return errors.New("not ready to log in")
	}
	_, err := conn.client.Login(user, pass)
	return err
}

// Open a new connection to an IMAP server.
func Connect(host string, user string, pass string) (*Connection, error) {
	conn := new(Connection)

	if err := conn.connect(host); err != nil {
		return nil, err
	}

	if err := conn.login(user, pass); err != nil {
		return nil, err
	}

	return conn, nil
}

// Log out from the IMAP server and close the connection.
func (conn *Connection) Close() {
	conn.client.Logout(3)
}

// Retrieve metadata for all messages in a given mailbox.
func (conn *Connection) Messages(box string, count uint32) ([]*MessageMeta,
	error) {
	conn.client.Select(box, true)

	// Create a range set that selects the most recent `count` messages (or
	// all messages if the mailbox contains fewer than that). Also, allocate a
	// slice for the results.
	set, _ := imap.NewSeqSet("")
	var messages []*MessageMeta
	totalCount := conn.client.Mailbox.Messages
	if totalCount >= count {
		set.AddRange(totalCount-(count-1), totalCount)
		messages = make([]*MessageMeta, count)
	} else {
		set.Add("1:*")
		messages = make([]*MessageMeta, totalCount)
	}

	// Fetch messages.
	cmd, err := conn.client.Fetch(set, "RFC822.HEADER RFC822.SIZE UID")
	if err != nil {
		return nil, err
	}

	// Parse each message.
	i := 0
	for cmd.InProgress() {
		err = conn.client.Recv(time.Second * 5)
		if err != nil {
			return nil, err
		}

		// Process the message data received so far.
		for _, rsp := range cmd.Data {
			msgInfo := rsp.MessageInfo()
			header := imap.AsBytes(msgInfo.Attrs["RFC822.HEADER"])
			msg, err := mail.ReadMessage(bytes.NewReader(header))
			if err != nil {
				return nil, err
			}

			messages[i] = &MessageMeta{
				UID:     msgInfo.UID,
				Size:    msgInfo.Size,
				Subject: msg.Header.Get("Subject"),
			}
			i++
		}
		cmd.Data = nil
	}

	return messages, nil
}
