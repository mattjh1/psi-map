package cli

import "github.com/urfave/cli/v2"

// serverCommand returns the server subcommand
func serverCommand() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"serve"},
		Usage:   "Start web server with analysis results",
		Flags:   serverFlags(),
		Action: func(c *cli.Context) error {
			return runAnalysis(c, true)
		},
	}
}

func serverFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "sitemap",
			Aliases:  []string{"s"},
			Usage:    "URL or sitemap.xml file path",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Server port",
			Value:   "8080",
		},
	}
}
