package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var markdownHeadingLinePattern = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

type markdownSegment struct {
	key string
	raw string
}

type markdownHeadingMatch struct {
	start int
	level int
	name  string
}

type markdownHeadingPathEntry struct {
	level int
	key   string
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func markdownSegments(content string) []markdownSegment {
	matches := markdownHeadingMatches(content)
	if len(matches) == 0 {
		return []markdownSegment{{key: "preamble", raw: content}}
	}
	segments := make([]markdownSegment, 0, len(matches)+1)
	if matches[0].start > 0 {
		segments = append(segments, markdownSegment{key: "preamble", raw: content[:matches[0].start]})
	}
	var path []markdownHeadingPathEntry
	for i, match := range matches {
		start := match.start
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1].start
		}
		key := markdownHeadingPathKey(match, &path)
		segments = append(segments, markdownSegment{key: key, raw: content[start:end]})
	}
	return segments
}

func markdownHeadingPathKey(match markdownHeadingMatch, path *[]markdownHeadingPathEntry) string {
	current := markdownHeadingSegmentKey(match.level, match.name)
	for len(*path) > 0 && (*path)[len(*path)-1].level >= match.level {
		*path = (*path)[:len(*path)-1]
	}
	parts := make([]string, 0, len(*path)+1)
	for _, entry := range *path {
		parts = append(parts, entry.key)
	}
	parts = append(parts, current)
	*path = append(*path, markdownHeadingPathEntry{level: match.level, key: current})
	return strings.Join(parts, " > ")
}

func markdownHeadingSegmentKey(level int, name string) string {
	return strings.Repeat("#", level) + " " + strings.ToLower(strings.TrimSpace(name))
}

func markdownHeadingMatches(content string) []markdownHeadingMatch {
	lines := strings.SplitAfter(content, "\n")
	var matches []markdownHeadingMatch
	offset := 0
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			offset += len(line)
			continue
		}
		if !inFence {
			withoutNewline := strings.TrimRight(line, "\r\n")
			parts := markdownHeadingLinePattern.FindStringSubmatch(withoutNewline)
			if len(parts) == 3 {
				matches = append(matches, markdownHeadingMatch{
					start: offset,
					level: len(parts[1]),
					name:  strings.TrimSpace(parts[2]),
				})
			}
		}
		offset += len(line)
	}
	return matches
}

func markdownSegmentMap(segments []markdownSegment) map[string]string {
	out := make(map[string]string, len(segments))
	for _, segment := range segments {
		out[segment.key] = segment.raw
	}
	return out
}
