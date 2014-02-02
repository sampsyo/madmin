package main

import (
    "github.com/sampsyo/madmin/mailbox"
    "fmt"
    "os"
    "log"
)

func trymail(host string, user string, pass string) {
    conn, err := mailbox.Connect(host, user, pass)
    if (err != nil) {
        log.Fatal("connection failed")
    }
    messages, err := conn.Messages("INBOX", 5)
    if (err != nil) {
        log.Fatal("failed to get messages", err.Error())
    }
    for _, mm := range messages {
        fmt.Println(mm.Subject)
    }
    conn.Close()
}

func main() {
    trymail(os.Args[1], os.Args[2], os.Args[3])
}
