package main

import (
	"code/crawler"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {
	var depth uint = 10
	var retries uint = 1
	var delay = 0 * time.Second
	var timeout = 15 * time.Second
	var rps uint
	var userAgent string
	var workers uint = 4
	var indentJSON bool

	cmd := &cli.Command{
		Name:  "hexlet-go-crawler",
		Usage: "analyze a website structure",
		Flags: []cli.Flag{
			&cli.UintFlag{Name: "depth", Usage: "crawl depth", Value: 10, Destination: &depth},
			&cli.UintFlag{Name: "retries", Usage: "number of retries for failed requests", Value: 1, Destination: &retries},
			&cli.DurationFlag{Name: "delay", Usage: "delay between requests (example: 200ms, 1s)", Value: 0 * time.Second, Destination: &delay},
			&cli.DurationFlag{Name: "timeout", Usage: "per-request timeout", Value: 15 * time.Second, Destination: &timeout},
			&cli.UintFlag{Name: "rps", Usage: "limit requests per second (overrides delay)", Value: 0, Destination: &rps},
			&cli.StringFlag{Name: "user-agent", Usage: "custom user agent", Destination: &userAgent},
			&cli.UintFlag{Name: "workers", Usage: "number of concurrent workers", Value: 4, Destination: &workers},
			&cli.BoolFlag{Name: "indent-json", Usage: "indent JSON output for readability without changing content or key order", Value: false, Destination: &indentJSON},
		},
		Commands: []*cli.Command{
			{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Shows a list of commands or help for one command",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("must specify at least one argument")
			}

			url := cmd.Args().First()
			if url == "" {
				return fmt.Errorf("must specify url")
			}

			options := crawler.Options{
				URL:         url,
				Depth:       depth,
				Retries:     retries,
				RPS:         rps,
				Delay:       delay,
				Timeout:     timeout,
				UserAgent:   userAgent,
				Concurrency: workers,
				IndentJSON:  indentJSON,
				HTTPClient:  &http.Client{Timeout: timeout},
			}

			res, err := crawler.Analyze(ctx, options)
			if err != nil {
				empty := crawler.Report{
					RootURL: url,
					Depth:   depth,
					Pages:   []crawler.Page{},
				}
				data, _ := json.Marshal(empty)
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(string(res))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(0)
	}
}
