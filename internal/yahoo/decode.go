package yahoo

import (
	"encoding/xml"
	"fmt"
	"io"
)

// APIError is Yahoo's error payload. It is returned both for error documents
// that arrive with HTTP 200 and, when parseable, for non-2xx responses.
type APIError struct {
	Description string `xml:"description"`
	Detail      string `xml:"detail"`
}

func (e *APIError) Error() string {
	if e.Detail == "" {
		return "yahoo: " + e.Description
	}
	return fmt.Sprintf("yahoo: %s: %s", e.Description, e.Detail)
}

// Envelope is the <fantasy_content> root every response shares. Exactly one of
// League, Players or Player is set, depending on the request.
type Envelope struct {
	XMLName xml.Name  `xml:"fantasy_content"`
	League  *League   `xml:"league"`
	Players *Players  `xml:"players"`
	Player  *Player   `xml:"player"`
	Error   *APIError `xml:"error"`
}

// decodeResponse decodes one response body. Yahoo reports failures as an
// <error> root, so error documents are returned as *APIError even with HTTP 200.
func decodeResponse(r io.Reader) (*Envelope, error) {
	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("yahoo: empty response")
			}
			return nil, err
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch start.Name.Local {
		case "error":
			var e APIError
			if err := dec.DecodeElement(&e, &start); err != nil {
				return nil, err
			}
			return nil, &e
		case "fantasy_content":
			var env Envelope
			if err := dec.DecodeElement(&env, &start); err != nil {
				return nil, err
			}
			return &env, nil
		default:
			return nil, fmt.Errorf("yahoo: unexpected root element <%s>", start.Name.Local)
		}
	}
}
