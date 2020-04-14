// @File ping.go
// @Author Issa Mathno(email@issa-mathno.com)
// @Description this is a simple CLI program that pings 
//              a given address. We are using a library
//              called go-ping by @Author sparrc
//              (https://github.com/sparrc/go-ping)
package main

import (
	"fmt"
        "os"
        "os/signal"
        "syscall"
        "regexp"
        "time"
        "github.com/sparrc/go-ping"
)

const REGEX_IP = "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"
const REGEX_URL = "^(http:\\/\\/www\\.|https:\\/\\/www\\.|http:\\/\\/|https:\\/\\/)?[a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,5}(:[0-9]{1,5})?(\\/.*)?$"

func main() {
        addr := ""
        if len(os.Args) >= 2 {
                addr = os.Args[1]
        } else {
                usage()
                addr = "www.cloudflare.com"
        }

        if (!is_valid_addr(addr)) {
                panic("The Address is invalid")
        }
        
        ping_addr(addr)
        os.Exit(0)
}

// This function will describe how to use the program and actually run the program
func usage() {
        fmt.Println("To use this program you will call the program and provide an appropriate IP address")
        fmt.Println("\tExample go run .\\src\\ping.go www.cloudflare.com")
        fmt.Println("\nThe Program will now ping Cloudflare's website until an interrupt is detected!\n")
        time.Sleep(5 * time.Second)
}

// This function uses regex to detect if the address is in the correct format
func is_valid_addr(addr string) bool {
        if (addr == "localhost") {
                return true
        }
        matched, _ := regexp.Match(REGEX_IP, []byte(addr))
        if (matched) {
                return true
        }

        matchedUrl, _ := regexp.Match(REGEX_URL, []byte(addr))
        return matchedUrl
}

// Pings the address with the windows native ping format with loss percentage on all packets
func ping_addr(addr string) {
        pinger, err := ping.NewPinger(addr)
        pinger.SetPrivileged(true) // Send "privelaged" raw ICMP pings. It mostly won't work without this
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

        fmt.Printf("Pinging %s with %v bytes of data:\n", addr, 32)

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