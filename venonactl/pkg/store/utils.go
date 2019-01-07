package store

const (
	Default = "master"
)

func GetLatestVersion() string {
	return Default
}
