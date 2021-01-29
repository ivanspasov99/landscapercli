// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"fmt"
	"html/template"
	"os"

	"github.com/gardener/landscaper/pkg/apis/core/v1alpha1"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const addManifestDeployItemUse = `deployitem \
    [component directory path] \
    [deployitem name] \
   `

const addManifestDeployItemExample = `
landscaper-cli component add manifest deployitem \
  . \
  nginx \
  --file ./deployment.yaml \
  --file ./service.yaml \
  --import-param target-ns(string)
`

const addManifestDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

//var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

type addManifestDeployItemOptions struct {
	componentPath  string
	deployItemName string

	files        *[]string
	importParams *[]string

	updateStrategy string
	policy         string

	clusterParam  string
	targetNsParam string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddManifestDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addManifestDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addManifestDeployItemUse,
		Example: addManifestDeployItemExample,
		Short:   addManifestDeployItemShort,
		Args:    cobra.ExactArgs(2),

		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Successfully added deploy item")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addManifestDeployItemOptions) Complete(args []string) error {
	o.componentPath = args[0]
	o.deployItemName = args[1]

	return o.validate()
}

func (o *addManifestDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	o.files = fs.StringArray(
		"file",
		[]string{},
		"manifest file")
	o.importParams = fs.StringArray(
		"import-param",
		[]string{},
		"import parameter")
	fs.StringVar(&o.updateStrategy,
		"update-strategy",
		"update",
		"update stategy")
	fs.StringVar(&o.policy,
		"policy",
		"",
		"policy")
	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"targetCluster",
		"target cluster")
	fs.StringVar(&o.targetNsParam,
		"target-ns-param",
		"",
		"target namespace")
}

func (o *addManifestDeployItemOptions) validate() error {
	if !identityKeyValidationRegexp.Match([]byte(o.deployItemName)) {
		return fmt.Errorf("the deploy item name must consist of lower case alphanumeric characters, '-', '_' " +
			"or '+', and must start and end with an alphanumeric character")
	}

	if o.targetNsParam == "" {
		return fmt.Errorf("target-ns-param is missing")
	}

	return nil
}

func (o *addManifestDeployItemOptions) run(ctx context.Context, log logr.Logger) error {
	blueprintPath := util.BlueprintDirectoryPath(o.componentPath)
	blueprint, err := blueprints.NewBlueprintReader(blueprintPath).Read()
	if err != nil {
		return err
	}

	if o.existsExecution(blueprint) {
		return fmt.Errorf("The blueprint already contains a deploy item %s\n", o.deployItemName)
	}

	exists, err := o.existsExecutionFile()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Deploy execution file %s already exists\n", util.ExecutionFilePath(o.componentPath, o.deployItemName))
	}

	err = o.createExecutionFile()
	if err != nil {
		return err
	}

	o.addExecution(blueprint)
	o.addImports(blueprint)
	return blueprints.NewBlueprintWriter(blueprintPath).Write(blueprint)
}

func (o *addManifestDeployItemOptions) existsExecution(blueprint *v1alpha1.Blueprint) bool {
	for i := range blueprint.DeployExecutions {
		execution := &blueprint.DeployExecutions[i]
		if execution.Name == o.deployItemName {
			return true
		}
	}

	return false
}

func (o *addManifestDeployItemOptions) addExecution(blueprint *v1alpha1.Blueprint) {
	blueprint.DeployExecutions = append(blueprint.DeployExecutions, v1alpha1.TemplateExecutor{
		Name: o.deployItemName,
		Type: v1alpha1.GOTemplateType,
		File: "/" + util.ExecutionFileName(o.deployItemName),
	})
}

func (o *addManifestDeployItemOptions) addImports(blueprint *v1alpha1.Blueprint) {
	o.addTargetImport(blueprint, o.clusterParam)
	o.addStringImport(blueprint, o.targetNsParam)
}

func (o *addManifestDeployItemOptions) addTargetImport(blueprint *v1alpha1.Blueprint, name string) {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == name {
			return
		}
	}

	required := true

	blueprint.Imports = append(blueprint.Imports, v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:       name,
			TargetType: string(v1alpha1.KubernetesClusterTargetType),
		},
		Required: &required,
	})
}

func (o *addManifestDeployItemOptions) addStringImport(blueprint *v1alpha1.Blueprint, name string) {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == name {
			return
		}
	}

	required := true

	blueprint.Imports = append(blueprint.Imports, v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   name,
			Schema: v1alpha1.JSONSchemaDefinition("{ \"type\": \"string\" }"),
		},
		Required: &required,
	})
}

func (o *addManifestDeployItemOptions) existsExecutionFile() (bool, error) {
	fileInfo, err := os.Stat(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if fileInfo.IsDir() {
		return false, fmt.Errorf("There already exists a directory %s\n", util.ExecutionFileName(o.deployItemName))
	}

	return true, nil
}

func (o *addManifestDeployItemOptions) createExecutionFile() error {
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		return err
	}

	defer f.Close()

	err = o.writeExecution(f)

	return err
}

const manifestExecutionTemplate = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/kubernetes-manifest
  target:
    name: {{"{{"}} .imports.{{.ClusterParam}}.metadata.name {{"}}"}}
    namespace: {{"{{"}} .imports.{{.ClusterParam}}.metadata.namespace {{"}}"}}
  config:
    apiVersion: manifest.deployer.landscaper.gardener.cloud/v1alpha2
    kind: ProviderConfiguration

    updateStrategy: {{.UpdateStrategy}}
	
	manifests: []
`

func (o *addManifestDeployItemOptions) writeExecution(f *os.File) error {
	t, err := template.New("").Parse(manifestExecutionTemplate)
	if err != nil {
		return err
	}

	data := struct {
		ClusterParam   string
		TargetNsParam  string
		DeployItemName string
		UpdateStrategy string
	}{
		ClusterParam:   o.clusterParam,
		TargetNsParam:  o.targetNsParam,
		DeployItemName: o.deployItemName,
		UpdateStrategy: o.updateStrategy,
	}

	err = t.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}
