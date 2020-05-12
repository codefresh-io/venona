package store

import (
	"context"
	"strings"

	"github.com/google/go-github/v21/github"
	version "github.com/hashicorp/go-version"
)

const (
	DefaultVersion = "latest"
)

type (
	logger interface {
		Debug(string, ...interface{})
		Error(string, ...interface{})
	}
)

// GetLatestVersion return the latest version of 0.x.x
// versions >= 1.0.0 are not compitable with 0.3.x and not installed using brew installation
// therefore, GetLatestVersion will return latest version in 0.x.x space
func GetLatestVersion(logger logger) string {
	defaultversion := DefaultVersion
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "codefresh-io", "venona", &github.ListOptions{})
	if err != nil {
		logger.Error("Request to get latest version of venona been rejected , setting version to latest. Original error: %s", err.Error())
		return defaultversion
	}
	for _, release := range releases {
		if release == nil {
			continue
		}
		if release.Draft == nil {
			continue
		}
		if *release.Draft {
			continue
		}
		name := strings.Split(*release.Name, "v")

		if len(name) < 2 {
			continue
		}
		latestAllowed, err := version.NewVersion("1.0.0")
		if err != nil {
			logger.Error("Failed to calc semver from version, setting version to latest. error: %s", err.Error())
			return DefaultVersion
		}
		current, err := version.NewVersion(name[1])
		if err != nil {
			logger.Error("Failed to calc semver from version, setting version to latest. error: %s", err.Error())
			return DefaultVersion
		}
		if current.LessThan(latestAllowed) {
			return current.String()
		}
	}
	return defaultversion
}

func IsRunningLatestVersion(s *Values) (bool, error) {
	current, err := version.NewVersion(s.Version.Current.Version)
	if err != nil {
		return false, err
	}
	latest, err := version.NewVersion(s.Version.Latest.Version)
	if err != nil {
		return false, err
	}
	if current.LessThan(latest) {
		return false, nil
	}
	return true, nil
}
