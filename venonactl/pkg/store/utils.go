package store

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
	version "github.com/hashicorp/go-version"
)

const (
	DefaultVersion = "latest"
)

func GetLatestVersion() string {
	version := DefaultVersion
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

func IsRunningLatestVersion() (bool, error) {
	s := GetStore()
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
