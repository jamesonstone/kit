package registry

import (
	"fmt"
	"sort"
)

func MergeSections(base []SectionHash, localContent, remoteContent string) (string, []string, error) {
	local, err := ParseMarkdown(localContent)
	if err != nil {
		return "", nil, fmt.Errorf("parse local artifact: %w", err)
	}
	remote, err := ParseMarkdown(remoteContent)
	if err != nil {
		return "", nil, fmt.Errorf("parse registry artifact: %w", err)
	}
	baseHashes := map[string]string{}
	for _, section := range base {
		baseHashes[section.Key] = section.Hash
	}
	keys := map[string]bool{}
	for key := range local.Sections {
		keys[key] = true
	}
	for key := range remote.Sections {
		keys[key] = true
	}
	merged := map[string]string{}
	var conflicts []string
	for key := range keys {
		localSection, localExists := local.Sections[key]
		remoteSection, remoteExists := remote.Sections[key]
		baseHash := baseHashes[key]
		localChanged := sectionChanged(localSection, localExists, baseHash)
		remoteChanged := sectionChanged(remoteSection, remoteExists, baseHash)
		switch {
		case localChanged && remoteChanged && localSection != remoteSection:
			conflicts = append(conflicts, key)
		case remoteChanged && remoteExists:
			merged[key] = remoteSection
		case !remoteChanged && localExists:
			merged[key] = localSection
		}
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return "", conflicts, nil
	}
	order := append([]string(nil), remote.Order...)
	order = append(order, local.Order...)
	return renderSections(order, merged), nil, nil
}

func sectionChanged(content string, exists bool, baseHash string) bool {
	if baseHash == "" {
		return exists
	}
	return !exists || HashContent(content) != baseHash
}
