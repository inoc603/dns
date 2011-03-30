package main

import (
        "os"
	"dns"
        "fmt"
)

func handleXfr(d *dns.Conn, i *dns.Msg) {
        if i.IsAxfr() {
                fmt.Printf("Axfr request seen\n")
                if i.Question[0].Name == Zone.name {
                        fmt.Printf("Matching current zone\n")
                        m := make(chan *dns.Xfr)
                        go d.XfrWrite(i, m)
                        for j := 0; j < Zone.size; j++ {
                                m <- &dns.Xfr{Add: true, RR: Zone.rrs[j]}
                        }
                        close(m)
                }
        }
}

func handleNotify(d *dns.Conn, i *dns.Msg) {
        if i.IsNotify() {
                fmt.Printf("Notify seen\n")
                q := new(dns.Msg)
                q.SetReply(i)
                err := d.WriteMsg(q)
                if err != nil {
                        return
                }
                doXfrIn(i)
        }
}

func doXfrIn(i *dns.Msg) ([]dns.RR, os.Error) {
        q := new(dns.Msg)
        q.SetAxfr(i.Question[0].Name)

        m := make(chan *dns.Xfr)
        fmt.Printf("Preparing Xfr for %s\n", i.Question[0].Name)

        d := new(dns.Conn)
        d.RemoteAddr = "127.0.0.1:53"
        err := d.Dial("tcp")
        if err != nil {
                return nil, err
        }
        defer d.Close()

        fmt.Printf("Calling 127.0.0.1 successful\n")
        go d.XfrRead(q, m)

        Zone.name = i.Question[0].Name
        j := 0
        for x := range m {
                fmt.Printf("%v %v\n", x.Add, x.RR)
                Zone.rrs[j] = x.RR
                j++
        }
        Zone.size = j
        return nil, nil
}
