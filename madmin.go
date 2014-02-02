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
    for mm := range conn.Messages("INBOX") {
        fmt.Println(mm)
    }
    conn.Logout()
}

func main() {
    trymail(os.Args[1], os.Args[2], os.Args[3])
}
