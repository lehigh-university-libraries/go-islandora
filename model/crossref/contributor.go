package crossref

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/lehigh-university-libraries/go-islandora/model"
	"github.com/lehigh-university-libraries/go-islandora/pkg/islandora"
	"golang.org/x/net/html"
)

type Contributor struct {
	Name         PersonName
	Suffix       string
	Organization string
	Orcid        string
	Role         string
	Sequence     string
}

type PersonName struct {
	Surname     string `xml:"surname"`
	Given       string `xml:"given_name,omitempty"`
	Institution string
	ORCID       string
}

func GetContributor(url string, first bool) Contributor {
	contributor := Contributor{
		Role: "author",
	}

	c, err := islandora.FetchTerm(url)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON for %s: %v", url, err)
	}

	name := c.Name.String()
	if name == "" {
		log.Fatalf("Bad term response: %s", url)
	}
	for _, r := range c.Relationships {
		if r.Url == "" {
			continue
		}
		if r.RelType != "schema:worksFor" {
			continue
		}

		relationshipUrl := fmt.Sprintf("https://preserve.lehigh.edu%s?_format=json", r.Url)
		respRel, err := http.Get(relationshipUrl)
		if err != nil {
			log.Fatalf("Error fetching relationship URL: %v", err)
		}
		defer respRel.Body.Close()

		if respRel.StatusCode != http.StatusOK {
			log.Fatalf("Error: received non-200 status code from relationships URL: %v", respRel.StatusCode)
		}

		bodyRel, err := io.ReadAll(respRel.Body)
		if err != nil {
			log.Fatalf("Error reading relationships response body: %v", err)
		}

		var cr model.TermResponse
		if err := json.Unmarshal(bodyRel, &cr); err != nil {
			log.Fatalf("Error unmarshaling relationships JSON: %v", err)
		}

		contributor.Name.Institution = cr.Name[0].Value
		name = strings.Replace(name, fmt.Sprintf(" - %s", contributor.Name.Institution), "", 1)
	}
	for _, i := range c.Identifier {
		if i.Attr0 != "orcid" {
			continue
		}

		contributor.Name.ORCID = i.Value
	}
	nameComponents := strings.Split(name, ", ")
	var surname string
	var given string
	if len(nameComponents) == 1 {
		nameComponents = strings.Split(name, " ")
		surname = nameComponents[len(nameComponents)-1]
		given = strings.Join(nameComponents[:len(nameComponents)-1], " ")
	} else {
		surname = nameComponents[0]
		given = ""
		if len(nameComponents) > 1 {
			given = strings.Join(nameComponents[1:], ", ")
		}

	}
	sequence := "additional"
	if first {
		sequence = "first"
	}
	contributor.Name.Given = html.EscapeString(given)
	contributor.Name.Surname = html.EscapeString(surname)
	contributor.Sequence = sequence
	return contributor
}
