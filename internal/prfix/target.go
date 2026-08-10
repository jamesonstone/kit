package prfix

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	githubURLPattern = regexp.MustCompile(`(?i)https?://github\.com/([^/\s)]+)/([^/\s)#]+)/pull/([1-9][0-9]*)`)
	ownerPRPattern   = regexp.MustCompile(`^([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+)#([1-9][0-9]*)$`)
	repositoryPart   = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

func ParseTarget(raw string, current Repository) (Target, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return Target{}, fmt.Errorf("pull request reference cannot be empty")
	}
	if match := githubURLPattern.FindStringSubmatch(value); match != nil {
		return targetFromParts(match[1], match[2], match[3])
	}
	if parsed, err := url.Parse(value); err == nil && strings.EqualFold(parsed.Host, "github.com") {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) == 4 && parts[2] == "pull" {
			return targetFromParts(parts[0], parts[1], parts[3])
		}
	}
	if match := ownerPRPattern.FindStringSubmatch(value); match != nil {
		return targetFromParts(match[1], match[2], match[3])
	}
	if number, err := strconv.Atoi(value); err == nil && number > 0 {
		if err := validateRepository(current); err != nil {
			return Target{}, fmt.Errorf("resolve current repository for PR %d: %w", number, err)
		}
		return Target{Repository: current, Number: number}, nil
	}
	return Target{}, fmt.Errorf("could not parse pull request reference %q", raw)
}

func targetFromParts(owner, name, numberText string) (Target, error) {
	repository := Repository{Owner: owner, Name: strings.TrimSuffix(name, ".git")}
	if err := validateRepository(repository); err != nil {
		return Target{}, err
	}
	number, err := strconv.Atoi(numberText)
	if err != nil || number < 1 {
		return Target{}, fmt.Errorf("invalid pull request number %q", numberText)
	}
	return Target{Repository: repository, Number: number}, nil
}

func validateRepository(repository Repository) error {
	if !safeRepositoryPart(repository.Owner) || !safeRepositoryPart(repository.Name) {
		return fmt.Errorf("invalid GitHub repository %q", repository.Slug())
	}
	return nil
}

func safeRepositoryPart(value string) bool {
	return value != "" && value != "." && value != ".." && repositoryPart.MatchString(value)
}

func SameRepository(left, right Repository) bool {
	return strings.EqualFold(left.Owner, right.Owner) && strings.EqualFold(left.Name, right.Name)
}
