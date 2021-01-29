package cmd

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var BaseContainer string
var appFS = afero.NewOsFs()

func init() {
	initCmd.Flags().StringVar(&BaseContainer, "base-container", "runiac:alpine-azure", "Base Docker image to use containing required tooling")

	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project",
	Long:  `Initialize a project`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = appFS.Mkdir(".runiac", 0755)

		dockerfile := strings.ReplaceAll(Dockerfile, "${BASE_CONTAINER}", BaseContainer)
		err := afero.WriteFile(appFS, ".runiac/Dockerfile", []byte(dockerfile), 0644)

		if err != nil {
			logrus.WithError(err).Error(err)
		}

		err = afero.WriteFile(appFS, ".runiac/.dockerignore", []byte(DockerIgnore), 0644)

		if err != nil {
			logrus.WithError(err).Error(err)
		}
	},
}

const DockerIgnore = `
.git
.runiac
.runiac/
`

const Dockerfile = `# do not edit --- autogenerated by runiac --- do not edit
# syntax = docker/dockerfile:experimental

ARG RUNIAC_CONTAINER="${BASE_CONTAINER}"

FROM golang:1.13 as builder

WORKDIR /app

RUN apt-get update && apt-get install -y unzip bash

RUN mkdir linux_amd64

SHELL ["/bin/bash", "-c"]

# Declare terraform providers and their versions here
RUN declare -A p; \
    p[random]=2.3.0; \
    p[archive]=1.2.2; \
    p[google]=3.51.1; \
    for provider in "${!p[@]}"; do version=${p[$provider]} && curl -o ${provider}.zip "https://releases.hashicorp.com/terraform-provider-${provider}/${version}/terraform-provider-${provider}_${version}_linux_amd64.zip" && unzip ${provider}.zip -d "/app/linux_amd64" && rm -f ${provider}.zip & done; wait;


FROM $RUNIAC_CONTAINER

COPY --from=builder /app/linux_amd64/ /root/.terraform.d/plugins/linux_amd64

WORKDIR /app

COPY . .

RUN mkdir -p $HOME/.terraform.d/plugin-cache

COPY entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh


ENTRYPOINT ["bash", "-c", "./entrypoint.sh"]
`
