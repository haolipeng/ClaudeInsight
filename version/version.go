package version

import (
	"claudeinsight/common"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

var (
	// String -X "claudeinsight/version.Version={{.Version}}"
	Version string
	// BuildTime -X "claudeinsight/version.CommitID={{.Commit}}"
	BuildTime string
	// CommitID -X "claudeinsight/version.BuildTime={{.Date}}"
	CommitID string
)

const unknown = "<unknown>"

func GetVersion() string {
	if Version != "" {
		return Version
	}
	return unknown
}

func GetBuildTime() string {
	if BuildTime != "" {
		return BuildTime
	}
	return unknown
}

func GetCommitID() string {
	if CommitID != "" {
		return CommitID
	}
	return unknown
}

func ReleaseVersion(ctx context.Context) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	client := github.NewClient(nil)
	releases, _, err := client.Repositories.GetLatestRelease(ctx, "hengyoush", "ClaudeInsight")
	if err != nil {
		return "", "", err
	}
	arch, err := common.UnameMachine()
	if err != nil {
		return "", "", err
	}
	ClaudeInsightAsset := fmt.Sprintf("ClaudeInsight_%s_linux_%s.tar.gz", strings.TrimPrefix(*releases.TagName, "v"), arch)
	for _, asset := range releases.Assets {
		if *asset.Name == ClaudeInsightAsset {
			return *releases.TagName, *asset.BrowserDownloadURL, nil
		}
	}
	return *releases.TagName, "", fmt.Errorf("no asset found for %s in github release", ClaudeInsightAsset)
}
