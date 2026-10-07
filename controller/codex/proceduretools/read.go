package proceduretools

import "encoding/json"

// ReadTexts changes only provider presentation. The recorded JSON observation
// stays intact, while source text is delivered without JSON string escapes.
func ReadTexts(value json.RawMessage) []string {
	var object map[string]json.RawMessage
	if json.Unmarshal(value, &object) != nil || object == nil {
		return []string{string(value)}
	}
	if batch, ok := object["results"]; ok {
		var results []json.RawMessage
		if json.Unmarshal(batch, &results) == nil {
			delete(object, "results")
			header, _ := json.Marshal(object)
			out := []string{string(header)}
			for _, r := range results {
				out = append(out, ReadTexts(r)...)
			}
			return out
		}
	}
	var text string
	if json.Unmarshal(object["text"], &text) != nil {
		return []string{string(value)}
	}
	delete(object, "text")
	metadata, _ := json.Marshal(object)
	return []string{string(metadata), text}
}
