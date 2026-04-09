package ui

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/benjamingriff/secretsrc/pkg/ui/components"
)

func parseSecretFields(value string) []components.SecretField {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &fields); err != nil {
		return nil
	}
	if len(fields) == 0 {
		return nil
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parsedFields := make([]components.SecretField, 0, len(keys))
	for _, key := range keys {
		copyValue, preview := prepareFieldCopyValue(fields[key])
		parsedFields = append(parsedFields, components.SecretField{
			Key:       key,
			CopyValue: copyValue,
			Preview:   preview,
		})
	}

	return parsedFields
}

func prepareFieldCopyValue(raw json.RawMessage) (string, string) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "null", "null"
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return stringValue, previewValue(stringValue)
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err == nil {
		value := compact.String()
		return value, previewValue(value)
	}

	value := string(raw)
	return value, previewValue(value)
}

func previewValue(value string) string {
	if value == "" {
		return "(empty)"
	}

	singleLine := strings.ReplaceAll(value, "\n", " ")
	if len(singleLine) <= 60 {
		return singleLine
	}

	return singleLine[:57] + "..."
}
