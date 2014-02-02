package mailbox

import (
    "code.google.com/p/go-imap/go1/imap"
    "log"
    "net/mail"
    "bytes"
    "errors"
)

type MessageMeta struct {
    UID     uint32
    Subject string
    Size    uint32
}

type Connection interface {
    Connect(host string) error
    Login(user string, pass string) error
    Messages(mailbox string) <-chan MessageMeta
    Logout()
}

type connection struct {
    client *imap.Client
}

func (conn *connection) Connect(host string) error {
    var err error
    conn.client, err = imap.DialTLS("imap.gmail.com", nil)
    if err != nil {
        return err
    }

    log.Print(conn.client.Data[0].Info)
    conn.client.Data = nil

    return nil
}

func (conn *connection) Login(user string, pass string) error {
    if conn.client.State() != imap.Login {
        return errors.New("not ready to log in")
    }
    _, err := conn.client.Login(user, pass)
    return err
}

func Connect(host string, user string, pass string) (Connection, error) {
    conn := new(connection)

    if err := conn.Connect(host); err != nil {
        return nil, err
    }

    if err := conn.Login(user, pass); err != nil {
        return nil, err
    }

    return conn, nil
}

func (conn *connection) Logout() {
    conn.client.Logout(3)
}

func getMessages(client *imap.Client, mailbox string,
                 messages chan<- MessageMeta) {
    client.Select(mailbox, true)

    // Choose first 10 messages
    set, _ := imap.NewSeqSet("")
    if client.Mailbox.Messages >= 10 {
        set.AddRange(client.Mailbox.Messages-9, client.Mailbox.Messages)
    } else {
        set.Add("1:*")
    }

    // Fetch messages
    cmd, _ := client.Fetch(set, "RFC822.HEADER RFC822.SIZE UID")
    for cmd.InProgress() {
        client.Recv(10)

        // Process returned messages
        for _, rsp := range cmd.Data {
            msgInfo := rsp.MessageInfo()
            header := imap.AsBytes(msgInfo.Attrs["RFC822.HEADER"])
            msg, _ := mail.ReadMessage(bytes.NewReader(header))

            mm := MessageMeta{
                UID: msgInfo.UID,
                Size: msgInfo.Size,
                Subject: msg.Header.Get("Subject"),
            }
            messages <- mm
        }
        cmd.Data = nil
    }

    close(messages)
}

func (conn *connection) Messages(mailbox string) <-chan MessageMeta {
    messages := make(chan MessageMeta, 10)
    go getMessages(conn.client, mailbox, messages)
    return messages
}
