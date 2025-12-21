package utils

import (
	"bufio"
	"droid-acp/types"
	"strings"
)

func GetPatchResult(patch string) (*types.PatchResult, error) {
	scanner := bufio.NewScanner(strings.NewReader(patch))
	result := &types.PatchResult{}
	var beforeLines []string
	var afterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if after, found := strings.CutPrefix(line, "*** Update File:"); found {
			result.URI = strings.TrimSpace(after)
			continue
		}
		if idx := strings.Index(line, "*** Update File:"); idx != -1 {
			result.URI = strings.TrimSpace(line[idx+len("*** Update File:"):])
			continue
		}
		if after, found := strings.CutPrefix(line, "Update File:"); found {
			result.URI = strings.TrimSpace(after)
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "***") ||
			strings.HasPrefix(trimmed, "@@") {
			continue
		}
		if after, found := strings.CutPrefix(line, "-"); found {
			beforeLines = append(beforeLines, after)
			continue
		}
		if after, found := strings.CutPrefix(line, "+"); found {
			afterLines = append(afterLines, after)
			continue
		}
	}
	result.Before = strings.Join(beforeLines, "\n")
	result.After = strings.Join(afterLines, "\n")
	return result, nil
}
