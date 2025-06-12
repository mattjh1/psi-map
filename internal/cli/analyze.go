package cli

import "github.com/urfave/cli/v2"

// analyzeCommand returns the analyze subcommand
func analyzeCommand() *cli.Command {
	return &cli.Command{
		Name:    "analyze",
		Aliases: []string{"run"},
		Usage:   "Analyze sitemap and generate reports as docs",
		Flags:   analyzeFlags(),
		Action: func(c *cli.Context) error {
			return runAnalysis(c, false)
		},
	}
}

func analyzeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "sitemap",
			Aliases:  []string{"s"},
			Usage:    "URL or sitemap.xml file path",
			Required: true,
		},
	}
}
