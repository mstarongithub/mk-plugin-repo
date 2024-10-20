package main

import "flag"

var (
	flagPrettyPrint = flag.Bool("pretty", false, "Tell the logger to pretty print to the console")
	flagLogLevel    = flag.String(
		"loglevel",
		"info",
		"Set the logging level to either debug, info, warn, error or fatal (case insensitive)",
	)
	flagInsecureCookies = flag.Bool(
		"insecureCookies",
		false,
		"Use insecure cookies for session storage",
	)
)

func init() {
	flag.Parse()
}
