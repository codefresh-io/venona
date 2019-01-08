package store

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
)

const (
	Default = "latest"
)

func getLatestVersion() string {
	version := Default
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "codefresh-io", "venona", &github.ListOptions{})
	if err != nil {
		logrus.Errorf("Request to get latest version of venona been rejected , setting version to latest. Original error: %s", err.Error())
		return version
	}
	for _, release := range releases {
		name := strings.Split(*release.Name, "v")
		if len(name) == 2 {
			return name[1]
		}
	}
	return version
}
