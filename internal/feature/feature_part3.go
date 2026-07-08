package feature

import "sort"

func DuplicateNumberGroups(features []Feature) map[int][]Feature {
	if len(features) == 0 {
		return nil
	}

	groups := make(map[int][]Feature)
	for _, feat := range features {
		groups[feat.Number] = append(groups[feat.Number], feat)
	}

	duplicates := make(map[int][]Feature)
	for number, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].DirName < group[j].DirName
		})
		duplicates[number] = group
	}

	if len(duplicates) == 0 {
		return nil
	}

	return duplicates
}
