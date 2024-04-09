package util

import "strings"

type RootUrlData struct {
	// The protocol specified. Usually http or https. Defaults to http if not set in original string
	Protocol string
	// Domain specified
	Domain string
	// The port specified. If not directly in the url, it will be infered from the protocol
	Port string
	// Whether the protocol used is "secure" (aka https) or not (http)
	IsSecure bool
}

// Helper func to take apart the RootUrl string in the Config
// Returns
func TakeApartRootUrlString(rootUrl string) RootUrlData {
	protocol := "http"
	domain := "localhost"
	port := "80"
	isSecure := false
	portSet := false
	front, back, found := strings.Cut(rootUrl, "://")

	if found {
		protocol = front
		front, back, found := strings.Cut(back, ":")
		if found {
			domain = front
			port = back
			portSet = true
		}
	} else {
		front, back, found := strings.Cut(front, ":")
		domain = front
		if found {
			port = back
			portSet = true
		}
	}
	if protocol == "https" {
		isSecure = true
		if !portSet {
			switch protocol {
			case "http":
				port = "80"
			case "https":
				port = "443"
			default:
				port = "80"
			}
		}
	}

	return RootUrlData{
		Protocol: protocol,
		Domain:   domain,
		Port:     port,
		IsSecure: isSecure,
	}
}
