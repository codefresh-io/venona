module github.com/codefresh-io/venona/venonactl

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/codefresh-io/go-sdk v0.18.0
	github.com/dustin/go-humanize v1.0.0
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/spf13/cobra v0.0.7
	github.com/spf13/viper v1.6.2
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v11.0.0+incompatible
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
