module github.com/codefresh-io/venona/venonactl

go 1.13

require (
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/codefresh-io/go-sdk v0.18.0
	github.com/dustin/go-humanize v1.0.0
	github.com/gophercloud/gophercloud v0.8.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm/v3 v3.1.1
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
