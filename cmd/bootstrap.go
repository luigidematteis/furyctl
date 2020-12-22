// Copyright (c) 2020 SIGHUP s.r.l All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sighupio/furyctl/internal/bootstrap"
	"github.com/sighupio/furyctl/internal/project"
	"github.com/sighupio/furyctl/pkg/analytics"
	"github.com/sighupio/furyctl/pkg/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func bPre(cmd *cobra.Command, args []string) (err error) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	handleStopSignal("bootstrap", stop)

	log.Debug("passing pre-flight checks")
	err = parseConfig(bConfigFilePath, "Bootstrap")
	if err != nil {
		log.Errorf("error parsing configuration file: %v", err)
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	workingDirFullPath := fmt.Sprintf("%v/%v", wd, bWorkingDir)
	log.Debug("pre-flight checks ok!")
	prj = &project.Project{
		Path: workingDirFullPath,
	}
	bootstrapOpts := &bootstrap.Options{
		Spin:                     s,
		Project:                  prj,
		ProvisionerConfiguration: cfg,
		TerraformOpts: &terraform.Options{
			GitHubToken: bGitHubToken,
			WorkingDir:  workingDirFullPath,
			Debug:       debug,
		},
	}
	boot, err = bootstrap.New(bootstrapOpts)
	if err != nil {
		log.Errorf("the bootstrap provisioner can not be initialized: %v", err)
		return err
	}
	return nil
}

var (
	boot *bootstrap.Bootstrap

	bConfigFilePath string
	bWorkingDir     string
	bGitHubToken    string
	bDryRun         bool

	bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Creates the required infrastructure to deploy a battle-tested Kubernetes cluster, mostly network components",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = cmd.Help()
			if err != nil {
				return err
			}
			return nil
		},
	}
	bootstrapInitCmd = &cobra.Command{
		Use:     "init",
		Short:   "Init a the project. Creates a directory with everything in place to apply the configuration",
		PreRunE: bPre,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			err = prj.Check()
			if err == nil {
				return fmt.Errorf("the project %v seems to be already created. Choose another working directory", bWorkingDir)
			}

			err = boot.Init()
			if err != nil {
				analytics.TrackBootstrapInit(bGitHubToken, false, cfg.Provisioner)
				return err
			}
			analytics.TrackBootstrapInit(bGitHubToken, true, cfg.Provisioner)
			return nil
		},
	}
	bootstrapUpdateCmd = &cobra.Command{
		Use:     "update",
		Short:   "Applies changes to the project. Running for the first time creates everything. Upcoming executions only applies changes.",
		PreRunE: bPre,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = prj.Check()
			if err != nil {
				return fmt.Errorf("the project %v has to be created. Execute bootstrap init before bootstrap update. %v", bWorkingDir, err)
			}

			err = boot.Update(bDryRun)
			if err != nil {
				analytics.TrackBootstrapUpdate(bGitHubToken, false, cfg.Provisioner, bDryRun)
				return err
			}
			analytics.TrackBootstrapUpdate(bGitHubToken, true, cfg.Provisioner, bDryRun)
			return nil
		},
	}
	bootstrapDestroyCmd = &cobra.Command{
		Use:     "destroy",
		Short:   "ATTENTION: Destroys the project. Ensure you destroy the cluster before destroying the bootstrap project.",
		PreRunE: bPre,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = prj.Check()
			if err != nil {
				return fmt.Errorf("the project %v has to be created. Execute bootstrap init before cluster destroy. %v", bWorkingDir, err)
			}

			err = boot.Destroy()
			if err != nil {
				analytics.TrackBootstrapDestroy(bGitHubToken, false, cfg.Provisioner)
				return err
			}
			analytics.TrackBootstrapDestroy(bGitHubToken, true, cfg.Provisioner)
			return nil
		},
	}
)

func init() {
	bootstrapUpdateCmd.PersistentFlags().BoolVar(&bDryRun, "dry-run", false, "Dry run execution")

	bootstrapInitCmd.PersistentFlags().StringVarP(&bConfigFilePath, "config", "c", "bootstrap.yml", "Bootstrap configuration file path")
	bootstrapUpdateCmd.PersistentFlags().StringVarP(&bConfigFilePath, "config", "c", "bootstrap.yml", "Bootstrap configuration file path")
	bootstrapDestroyCmd.PersistentFlags().StringVarP(&bConfigFilePath, "config", "c", "bootstrap.yml", "Bootstrap configuration file path")

	bootstrapInitCmd.PersistentFlags().StringVarP(&bWorkingDir, "workdir", "w", "./bootstrap", "Working directory to create and place all project files. Must not exists.")
	bootstrapUpdateCmd.PersistentFlags().StringVarP(&bWorkingDir, "workdir", "w", "./bootstrap", "Working directory with all project files")
	bootstrapDestroyCmd.PersistentFlags().StringVarP(&bWorkingDir, "workdir", "w", "./bootstrap", "Working directory with all project files")

	bootstrapInitCmd.PersistentFlags().StringVarP(&bGitHubToken, "token", "t", "", "GitHub token to access enterprise repositories. Contact sales@sighup.io")
	bootstrapUpdateCmd.PersistentFlags().StringVarP(&bGitHubToken, "token", "t", "", "GitHub token to access enterprise repositories. Contact sales@sighup.io")
	bootstrapDestroyCmd.PersistentFlags().StringVarP(&bGitHubToken, "token", "t", "", "GitHub token to access enterprise repositories. Contact sales@sighup.io")

	bootstrapCmd.AddCommand(bootstrapInitCmd)
	bootstrapCmd.AddCommand(bootstrapUpdateCmd)
	bootstrapCmd.AddCommand(bootstrapDestroyCmd)
	rootCmd.AddCommand(bootstrapCmd)
}
