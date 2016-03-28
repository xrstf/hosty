// Code in this file is based on the gosimple/slug package, but modified
// to allow dots in slugs. License of the original code is MPLv2.
// See https://github.com/gosimple/slug

package main

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/rainycape/unidecode"
)

var (
	regexpNonAuthorizedChars = regexp.MustCompile(`[^a-z0-9-_.]`)
	regexpMultipleDashes     = regexp.MustCompile(`-+`)
	regexpMultipleDots       = regexp.MustCompile(`\.+`)

	enSub = map[rune]string{
		'&':  " and ",
		'@':  " at ",
		'"':  "",
		'\'': "",
		'’':  "",
		'‒':  "-", // figure dash
		'–':  "-", // en dash
		'—':  "-", // em dash
		'―':  "-", // horizontal bar
	}
)

func makeSlug(s string, maxLen int) (slug string) {
	slug = strings.TrimSpace(s)
	slug = substituteRune(slug, enSub)

	// Process all non ASCII symbols
	slug = unidecode.Unidecode(slug)
	slug = strings.ToLower(slug)

	// Process all remaining symbols
	slug = regexpNonAuthorizedChars.ReplaceAllString(slug, "-")
	slug = regexpMultipleDashes.ReplaceAllString(slug, "-")
	slug = regexpMultipleDots.ReplaceAllString(slug, ".")
	slug = strings.Trim(slug, ".-")

	if maxLen > 0 {
		slug = smartTruncate(slug, maxLen)
		slug = strings.Trim(slug, ".-")
	}

	return slug
}

func substituteRune(s string, sub map[rune]string) string {
	var buf bytes.Buffer
	for _, c := range s {
		if d, ok := sub[c]; ok {
			buf.WriteString(d)
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func smartTruncate(text string, maxLen int) string {
	if len(text) < maxLen {
		return text
	}

	var truncated string
	words := strings.SplitAfter(text, "-")
	// If maxLen is smaller than length of the first word return word
	// truncated after maxLen.
	if len(words[0]) > maxLen {
		return words[0][:maxLen]
	}
	for _, word := range words {
		if len(truncated)+len(word)-1 <= maxLen {
			truncated = truncated + word
		} else {
			break
		}
	}

	return truncated
}
