package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// URI is the structured form of host/resource-group</site>?opts
type URI struct {
	Host          string
	ResourceGroup string `json:",omitempty"`
	Site          string `json:",omitempty"`
	Insecure      bool
}

// ParseURI will parse a given string into a URI
func ParseURI(value string) (*URI, error) {
	var opts url.Values
	var err error
	if !strings.HasPrefix(value, "//") {
		return nil, fmt.Errorf("invalid URI; must begin with //")
	}
	if strings.Contains(value, "?") {
		parts := strings.Split(value, "?")
		value = parts[0]
		if opts, err = url.ParseQuery(parts[1]); err != nil {
			return nil, err
		}
	}
	var insecure bool
	insecure, _ = strconv.ParseBool(opts.Get("insecure"))
	parts := strings.Split(value[2:], "/")
	var uri URI
	if len(parts) == 1 {
		uri = URI{
			Host:     parts[0],
			Insecure: insecure,
		}
	} else if len(parts) == 2 {
		uri = URI{
			Host:          parts[0],
			ResourceGroup: parts[1],
			Insecure:      insecure,
		}
	} else if len(parts) == 3 {
		uri = URI{
			Host:          parts[0],
			ResourceGroup: parts[1],
			Site:          parts[2],
			Insecure:      insecure,
		}
	} else {
		return nil, fmt.Errorf("invalid format")
	}
	return &uri, nil
}

// Equals will test equality of two URIs
func (uri *URI) Equals(other *URI) bool {
	return uri.Host == other.Host && uri.ResourceGroup == other.ResourceGroup && uri.Site == other.Site
}

func (uri *URI) String() string {
	if uri.ResourceGroup == "" {
		return fmt.Sprintf("//%s", uri.Host)
	}
	if uri.Site == "" {
		return fmt.Sprintf("//%s/%s", uri.Host, uri.ResourceGroup)
	}
	return fmt.Sprintf("//%s/%s/%s", uri.Host, uri.ResourceGroup, uri.Site)
}

// GetURIFromArg will parse a string looking for a pattern to build a URI. The
// values may come from the argument or from the default cluster.
func GetURIFromArg(arg string) (*URI, error) {
	uri, err := ParseURI(arg)
	if err != nil {
		if config.DefaultCluster == nil {
			return nil, errors.New("first argument must be a URI or a default cluster must be set")
		}
		if arg == "" {
			return config.DefaultCluster, nil
		}
		switch strings.Count(arg, "/") {
		case 0:
			if config.DefaultCluster.ResourceGroup == "" {
				// arg is the resource group
			} else {
				// arg is the site
			}
			break
		}
	}
	return uri, nil
}
