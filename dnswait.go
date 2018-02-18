package main

import (
  "github.com/0xAX/notificator"
  spin "github.com/briandowns/spinner"
  "github.com/urfave/cli"

  "fmt"
  "log"
  "net"
  "net/url"
  "os"
  "time"
)

var Options = &struct {
  Domain              string
  IpAddress           string
  Time                int64
  DisableNotification bool
}{}

func main() {
  app := &cli.App{
    Name:        "dnswait",
    Version:     "0.0.1",
    Author:      "Stefan Ruzitschka",
    Description: "Waits for given domain to resolve to a given IP.\n   A notification will be sent, when the program ends or the domain successfully resolves to the IP.\n\n   Supported notifications:\n   - Linux: notify-send or kdialog \n   - OSX: terminal-notifier or osascript \n   - Windows: growlnotify",
    UsageText:   "dnswait --domain <domain> --ip <ipv4|ipv6> [--time <minutes>] [--disable-notification]\n\n   dnswait --domain ipv6.google.com --ip 2a00:1450:4002:80b::200e\n   dnswait --domain reddit.com --ip 151.101.65.140",
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:        "domain",
        Destination: &Options.Domain,
        Usage:       "domain to resolve",
        Value:       "",
      },
      &cli.StringFlag{
        Name:        "ip",
        Destination: &Options.IpAddress,
        Usage:       "target ip (IPv4 | IPv6)",
      },
      &cli.Int64Flag{
        Name:        "time",
        Destination: &Options.Time,
        Usage:       "time to wait in minutes",
        Value:       20,
      },
      &cli.BoolFlag{
        Name:        "disable-notification",
        Destination: &Options.DisableNotification,
        Usage:       "disables notification",
      },
    },
  }

  app.Action = func(context *cli.Context) error {
    if !context.IsSet("domain") {
      log.Fatalln("Please set a domain")
    }

    if !context.IsSet("ip") {
      log.Fatalln("Please set an IP")
    }

    var ip = net.ParseIP(Options.IpAddress)

    if ip == nil {
      log.Fatalf("Cannot parse ip %s", Options.IpAddress)
    }

    domain, err := url.Parse(Options.Domain)

    if err != nil {
      log.Println(err)

      os.Exit(1)
    }

    var duration = 20 * time.Minute

    if Options.Time > 0 {
      duration = time.Duration(Options.Time) * time.Minute
    }

    spinner := spin.New(spin.CharSets[25], 100*time.Millisecond)
    spinner.Start()
    spinner.Suffix = fmt.Sprintf(" Waiting %s for %s to point to %s...", duration, domain, ip.String())

    ticker := time.NewTicker(5 * time.Second)

    notify := notificator.New(notificator.Options{
      AppName: "dnswait",
    })

    go func() {
      for _ = range ticker.C {
        ips, err := net.LookupHost(domain.String())

        if err != nil {
          log.Println(err)

          os.Exit(1)
        }

        for _, foundIp := range ips {
          if foundIp == ip.String() {
            if !Options.DisableNotification {
              notify.Push(
                "dnswait succeeded!",
                fmt.Sprintf("%s resolves to %s!", domain, ip.String()),
                "",
                notificator.UR_NORMAL,
              )
            }

            os.Exit(0)
          }
        }
      }
    }()

    <-time.After(duration)

    ticker.Stop()
    spinner.Stop()

    if !Options.DisableNotification {
      notify.Push("dnswait failed!",
        fmt.Sprintf("%s does not resolve to %s!", domain, ip.String()),
        "",
        notificator.UR_CRITICAL,
      )
    }

    return cli.NewExitError("", 1)
  }

  app.Run(os.Args)
}
