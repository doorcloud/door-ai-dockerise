package xmlutil

import (
	"encoding/xml"
	"strings"
)

// FindText walks an encoding/xml.Decoder until the given path
// (e.g., parent -> version) is found and returns its text.
func FindText(dec *xml.Decoder, path ...string) (string, bool) {
	if len(path) == 0 {
		return "", false
	}

	var currentPath []string
	var foundText string

	for {
		token, err := dec.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			currentPath = append(currentPath, t.Name.Local)
		case xml.EndElement:
			if len(currentPath) > 0 {
				currentPath = currentPath[:len(currentPath)-1]
			}
		case xml.CharData:
			if len(currentPath) == len(path) {
				match := true
				for i, p := range path {
					if currentPath[i] != p {
						match = false
						break
					}
				}
				if match {
					foundText = strings.TrimSpace(string(t))
					return foundText, true
				}
			}
		}
	}

	return "", false
}
