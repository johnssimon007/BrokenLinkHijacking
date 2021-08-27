package main

import (
   "crypto/tls"
   "flag"
   "fmt"
   "github.com/common-nighthawk/go-figure"
   "github.com/jackdanger/collectlinks"
   "net"
   "net/http"
   "net/url"
   "os"
   "strings"
   "sync"
   "time"
)

var visited = sync.Map{}

func main() {
   ascii := figure.NewColorFigure("Broken Link Hijacker", "", "yellow", true)
   ascii.Print()
   flag.Parse()

   args := flag.Args()
   fmt.Println(args)
   if len(args) < 1 {
      fmt.Println("Please specify a domain in the format specified below, \n Usage Example: go run main.go https://example.com")
      os.Exit(1)
   }

   queue := make(chan string)

   go func() {
      queue <- args[0]
   }()

   fmt.Printf(string("\033[1;33m [X] Starting Crawler -  This May Take Some Time Depending Upon The Amount Of Links Present \n \033[0m"))

   for uri := range queue {
      go func() {
         enqueue(uri, queue)
      }()
   }
}

func enqueue(uri string, queue chan string) {
   visited.Store(uri, true)

   transport := &http.Transport{
      MaxIdleConns:      30,
      IdleConnTimeout:   time.Second,
      DisableKeepAlives: true,
      TLSClientConfig: &tls.Config{
         InsecureSkipVerify: true,
      },
      DialContext: (&net.Dialer{
         Timeout:   2 * time.Second,
         KeepAlive: time.Second,
      }).DialContext,
   }

   client := http.Client{
      Transport: transport,
      Timeout:   2 * time.Second,
   }
   resp, err := client.Get(uri)

   if err != nil {
      return

   }
   if resp != nil {
      if resp.Body != nil {
         defer resp.Body.Close()
      }
   }

   links := collectlinks.All(resp.Body)
   status_codes := map[int]string{
      404: "Resource Not FOUND",
   }
   domain_list := map[string]string{
      "linkedin.com":  "Linkedin",
      "facebook.com":  "Facebook",
      "twittter.com":  "twitter",
      "youtube.com":   "youtube",
      "twitch.com":    "twitch",
      "discord.com":   "discord",
      "instagram.com": "instagram",
   }

   for _, link := range links {
      absolute := fixUrl(link, uri)
      if uri != "" {
         _, already_visited := visited.Load(absolute)
         if !already_visited {
            response, error := client.Get(absolute)

            if error != nil {
               return
            }
            if response != nil {
               if response.Body != nil {
                  defer response.Body.Close()
               }
            }

            u, err := url.Parse(absolute)
            if err != nil {
               return
            }

            parts := strings.Split(u.Hostname(), ".")
            domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
            _, exists := status_codes[response.StatusCode]

            _, domain_exists := domain_list[domain]

            if exists && domain_exists {
               fmt.Printf(string("\033[1;33m %s Seems to be vulnerable\033[0m \n"), absolute)
            } else if exists && !domain_exists {
               fmt.Printf(string("\033[1;33m %s Might be vulnerable\033[0m \n"), absolute)
            }
        
            fmt.Println(absolute)

            go func() {
               queue <- absolute
            }()
         }
      }
   }
}

func fixUrl(href, base string) string {
   uri, err := url.Parse(href)
   if err != nil {
      return ""
   }
   baseUrl, err := url.Parse(base)
   if err != nil {
      return ""
   }
   uri = baseUrl.ResolveReference(uri)
   return uri.String()
}
