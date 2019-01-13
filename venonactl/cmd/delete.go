package cmd

/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"errors"
	"os"

	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/sirupsen/logrus"

	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/spf13/cobra"
)

type DeletionError struct {
	err       error
	operation string
	name      string
}

// deleteCmd represents the status command
var deleteCmd = &cobra.Command{
	Use:   "delete [names]",
	Short: "Delete a Codefresh's runtime-environment",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires name of the runtime-environment")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()
		var errors []DeletionError
		contextName := ""
		kubeContextFlag := cmd.Flag("kube-context-name")
		if kubeContextFlag != nil {
			contextName = kubeContextFlag.Value.String()
		}
		for _, name := range args {
			re, err := s.CodefreshAPI.Client.RuntimeEnvironments().Get(name)
			if err != nil {
				errors = append(errors, DeletionError{
					err:       err,
					name:      name,
					operation: "Get Runtime-Environment from Codefresh",
				})
				continue
			}
			deleted, err := s.CodefreshAPI.Client.RuntimeEnvironments().Delete(name)

			if err != nil {
				errors = append(errors, DeletionError{
					err:       err,
					name:      name,
					operation: "Delete Runtime-Environment from Codefresh",
				})
				continue
			}

			if deleted {
				if contextName == "" {
					contextName = re.RuntimeScheduler.Cluster.ClusterProvider.Selector
				}
				s.KubernetesAPI.ContextName = contextName
				s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace
				err = runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Delete()
				if err != nil {
					errors = append(errors, DeletionError{
						err:       err,
						name:      name,
						operation: "Delete Runtime-Environment Kubernetes resoruces",
					})
					continue
				}
				if re.Metadata.Agent {
					err = runtimectl.GetOperator(runtimectl.VenonaOperatorType).Delete()
					if err != nil {
						errors = append(errors, DeletionError{
							err:       err,
							name:      name,
							operation: "Delete Venona's agent Kubernetes resoruces",
						})
						continue
					}
				}
				logrus.Infof("Deleted %s", name)
			}

		}

		if len(errors) > 0 {
			for _, e := range errors {
				logrus.WithFields(logrus.Fields{
					"runtime-environment": e.name,
				}).Errorf("Failed during operation %s with error %s", e.operation, e.err.Error())
			}
			os.Exit(1)
		}
		logrus.Info("Deletion completed")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().String("kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
}
