package main

import (
	"fmt"
        "github.com/sparrc/go-ping"
        "os"
        "os/signal"
        "syscall"
        "regexp"
)

const REGEX_IP = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"
const REGEX_URL = "^(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?$"

func main() {
        ip := ""
        if len(os.Args) >= 2 {
                ip = os.Args[1]
        } else {
                ip = "www.cloudflare.com"
        }

        if (!is_valid_addr(ip)) {
                panic("The Address is invalid")
        }
        ping_addr(ip)
        os.Exit(0)
}

// This function uses regex to detect if the ip is in the correct format
func is_valid_addr(ip string) bool {
        if (ip == "localhost") {
                return true
        }
        matched, _ := regexp.Match(REGEX_IP, []byte(ip))
        if (matched) {
                return true
        }

        matchedUrl, _ := regexp.Match(REGEX_URL, []byte(ip))
        return matchedUrl
}

func ping_addr(ip string) {
        pinger, err := ping.NewPinger(ip)
        pinger.SetPrivileged(true) // Send "privelaged" raw ICMP pings
        pinger.Size = 24 // Mimicks the 32 bytes that is like the CMD for windows.

        if err != nil {
                panic(err)
        }
        
        // Interrupt; Quit the function
        c := make(chan os.Signal)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        go func() {
                <-c
                pinger.Stop()
        }()

        fmt.Printf("Pinging %s with %v bytes of data:\n", ip, 32)

        pinger.OnRecv = func(pkt *ping.Packet) {
                stats := pinger.Statistics()
                fmt.Printf("Reply from %s bytes=%d %v%% loss time=%v\n",
                        pkt.IPAddr, pkt.Nbytes, stats.PacketLoss, pkt.Rtt)
        }

        pinger.OnFinish = func(stats *ping.Statistics) {
                fmt.Printf("Ping statistics for %s:\n", stats.IPAddr)
                fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%v%% loss)\n",
                        stats.PacketsSent, stats.PacketsRecv, stats.PacketsSent - stats.PacketsRecv, stats.PacketLoss)
                fmt.Println("Approximate round trip times in milli-seconds:")
                fmt.Printf("\tMinimum = %v, Maximum = %v, Average = %v", stats.MinRtt, stats.MaxRtt, stats.AvgRtt)
        }

        pinger.Run()
}