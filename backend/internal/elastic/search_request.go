package elastic

import (
	"encoding/json"
	"fmt"
)

const (
	searchHighlightPreTag          = "<mark>"
	searchHighlightPostTag         = "</mark>"
	searchHighlightFragmentSize    = 300
	searchHighlightFragmentsNumber = 1
)

func buildSearchRequestBody(query string, page int, limit int) ([]byte, error) {
	from := (page - 1) * limit

	searchBody := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"text"},
			},
		},
		"from": from,
		"size": limit,
		"highlight": map[string]any{
			"pre_tags":  []string{searchHighlightPreTag},
			"post_tags": []string{searchHighlightPostTag},
			"fields": map[string]any{
				"text": map[string]any{
					"fragment_size":       searchHighlightFragmentSize,
					"number_of_fragments": searchHighlightFragmentsNumber,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("marshal search body: %w", err)
	}

	return bodyBytes, nil
}
