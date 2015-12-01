package main

import (
	"strings"
	"unicode"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	gogo "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
	"github.com/serenize/snaker"
)

func main() {
	req := command.Read()
	files := req.GetProtoFile()

	// Forced (to prevent mistake)
	vanity.ForEachFile(files, vanity.TurnOffGogoImport)

	// Default stuffs.
	vanity.ForEachFile(files, vanity.TurnOnSizerAll) // 'Marshal' needs this.
	vanity.ForEachFile(files, vanity.TurnOnUnmarshalerAll)
	vanity.ForEachFile(files, vanity.TurnOnMarshalerAll)

	// Apply custom fixes.
	vanity.ForEachFieldInFilesExcludingExtensions(files, fixFieldName)
	// vanity.ForEachFieldInFilesExcludingExtensions(files, fixNumericJSONTag)

	resp := command.Generate(req)
	command.Write(resp)
}

// fixFieldName changes field names like
//  'id' -> 'ID'
func fixFieldName(field *gogo.FieldDescriptorProto) {
	if gogoproto.IsCustomName(field) {
		return // Skip if a custom name is specified.
	}

	if field.Options == nil {
		field.Options = new(gogo.FieldOptions)
	}

	name := field.GetName()
	name = snaker.SnakeToCamel(name)
	name = lintName(name)

	// Use gogoproto.customname
	proto.SetExtension(field.Options, gogoproto.E_Customname, &name)
}

//TODO(jerrodrurik):
// func fixNumericJSONTag(field *gogo.FieldDescriptorProto) {
// 	if gogoproto.GetJsonTag(field) != nil {
// 		return // Skip if a custom jsontag is specified.
// 	}
//
// 	switch field.GetType() {
// 	// https://godoc.org/github.com/golang/protobuf/jsonpb says
// 	// jsonpb "encodes int64, uint64 as strings"
// 	case gogo.FieldDescriptorProto_TYPE_INT64, gogo.FieldDescriptorProto_TYPE_UINT64:
//
// 	default: // Not applicable.
// 		return
// 	}
//
// 	if field.Options == nil {
// 		field.Options = new(gogo.FieldOptions)
// 	}
//
// 	jsonTag := *field.Name + ",string" + ",omitempty"
// 	proto.SetExtension(field.Options, gogoproto.E_Jsontag, &jsonTag)
// }

// See: https://github.com/golang/lint/blob/master/lint.go
// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XSRF":  true,
	"XSS":   true,
}

// lintName returns a different name if it should be different.
func lintName(name string) (should string) {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}
