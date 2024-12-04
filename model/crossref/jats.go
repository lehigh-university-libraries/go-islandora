package crossref

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html"
)

type Body struct {
	Sections []Section `xml:"sec"`
}

type Section struct {
	XMLName xml.Name `xml:"jats:abstract"`
	Para    []string `xml:"jats:p"`
}

func parseHTML(htmlContent string) (*html.Node, error) {
	doc, err := html.Parse(bytes.NewReader([]byte(htmlContent)))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func StrToJATS(htmlContent string) (string, error) {
	if len(htmlContent) > 4 && htmlContent[0:4] == "</p>" {
		htmlContent = "<p>" + htmlContent[4:]
	}
	node, err := parseHTML(htmlContent)
	if err != nil {
		return "", err
	}

	var section Section
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			section.Para = append(section.Para, n.FirstChild.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)

	output, err := xml.Marshal(section)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
