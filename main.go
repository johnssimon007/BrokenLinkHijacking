package main

import (
    "crypto/tls"
    "flag"
    "fmt"
    "github.com/jackdanger/collectlinks"
    "net/http"
    "github.com/gookit/color"
    "strings"
    "net/url"
    "os"
)

var sem = make(chan struct{}, 20)

func usage() {
    fmt.Fprintf(os.Stderr, "usage: BrokenLinkHijacking https://example.com\n")
    flag.PrintDefaults()
    os.Exit(2)
}

func main() {
    flag.Usage = usage
    flag.Parse()

    args := flag.Args()
    fmt.Println(args)
    if len(args) < 1 {
        usage()
        fmt.Println("Please specify start page")
        os.Exit(1)
    }

    queue := make(chan string)
    filteredQueue := make(chan string)

    go func() { queue <- args[0] }()
    go filterQueue(queue, filteredQueue)

    done := make(chan bool)

    for i := 0; i < 150; i++ {
        go func() {
            sem <- struct{}{}
            for uri := range filteredQueue {
                enqueue(uri, queue)
            }
            <-sem
            done <- true
        }()
    }
    <-done
}

func filterQueue(in chan string, out chan string) {
    var seen = make(map[string]bool)
    for val := range in {
        if !seen[val] {
            seen[val] = true
            out <- val
        }
    }
}

func enqueue(uri string, queue chan string) {
    fmt.Println("fetching", uri)
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    }
    client := http.Client{Transport: transport}
    resp, err := client.Get(uri)
    if err != nil {
        return
    }
    defer resp.Body.Close()

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
					 color.Danger.Println("Seems to be vulnerable "+ absolute)
         } else if exists && !domain_exists {
					 color.Info.Println("Might be vulnerable "+absolute)

         }


			go func() { queue <- absolute }()
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
