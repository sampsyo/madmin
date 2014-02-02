// A simple IMAP email client that abstracts away the details of the IMAP
// protocol.
package mailbox

import (
    "code.google.com/p/go-imap/go1/imap"
    "net/mail"
    "bytes"
    "errors"
)

// A connection to an IMAP server.
type Connection struct {
    client *imap.Client
}

// The metadata for an email message. Includes "envelope" information but not
// the body of the message.
type MessageMeta struct {
    // The message's unique identifier.
    UID     uint32

    // The subject line.
    Subject string

    // The message's total size in bytes.
    Size    uint32
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

func getMetadata(client *imap.Client, box string, count uint32,
                 messages chan<- MessageMeta) {
    client.Select(box, true)

    // Create a range set that selects the most recent `count` messages (or
    // all messages if the mailbox contains fewer than that).
    set, _ := imap.NewSeqSet("")
    if client.Mailbox.Messages >= count {
        set.AddRange(client.Mailbox.Messages - (count - 1),
                     client.Mailbox.Messages)
    } else {
        set.Add("1:*")
    }

    // Fetch messages
    // FIXME handle errors gracefully
    cmd, _ := client.Fetch(set, "RFC822.HEADER RFC822.SIZE UID")
    for cmd.InProgress() {
        // FIXME handle timeout/error
        client.Recv(10)

        // Process the message data received so far.
        for _, rsp := range cmd.Data {
            msgInfo := rsp.MessageInfo()
            header := imap.AsBytes(msgInfo.Attrs["RFC822.HEADER"])
            msg, _ := mail.ReadMessage(bytes.NewReader(header))

            messages <- MessageMeta{
                UID: msgInfo.UID,
                Size: msgInfo.Size,
                Subject: msg.Header.Get("Subject"),
            }
        }
        cmd.Data = nil
    }

    close(messages)
}

// Retrieve metadata for all messages in a given mailbox. Messages are
// requested asynchronously, but be sure to consume all messages before
// issuing any more commands.
func (conn *Connection) Messages(box string, count uint32) <-chan MessageMeta {
    messages := make(chan MessageMeta, 10)
    go getMetadata(conn.client, box, count, messages)
    return messages
}
