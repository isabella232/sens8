package check

import (
	"fmt"
	"encoding/json"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

type DeploymentStatus struct {
	BaseCheck
	warnLevel   *float32
	critLevel   *float32
	deployment  *extensions.Deployment
	commandLine *flag.FlagSet
}

//NewDeploymentStatus creates a new deployment health check
func NewDeploymentStatus(config CheckConfig) (Check, error) {
	dh := DeploymentStatus{}
	dh.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	dh.warnLevel = commandLine.Float32P("warn", "w", 0.9, "Percent of healthy pods to warn at")
	dh.critLevel = commandLine.Float32P("crit", "c", 0.8, "Percent of healthy pods to alert critical at")
	err := commandLine.Parse(config.Argv[1:])
	dh.commandLine = commandLine
	if err != nil {
		return &dh, err
	}
	if *dh.warnLevel <= float32(0) || *dh.warnLevel > float32(1) {
		return &dh, fmt.Errorf("--warn must be > 0 and <= 1")
	}
	if *dh.critLevel <= float32(0) || *dh.critLevel > float32(1) {
		return &dh, fmt.Errorf("--cirt must be > 0 and <= 1")
	}

	return &dh, nil
}

func (dh *DeploymentStatus) Usage() CheckUsage {
	return CheckUsage{
		Description: `Checks deployment pod levels via status obj given by Kubernetes. Provides full deployment status object in result output`,
		Flags: dh.commandLine.FlagUsages(),
	}
}

func (dh *DeploymentStatus) Update(resource interface{}) {
	dh.deployment = resource.(*extensions.Deployment)
}

func (dh *DeploymentStatus) Execute() (CheckResult, error) {
	res := NewCheckResultFromConfig(dh.Config)

	status := dh.deployment.Status

	res.Status = OK
	level := float32(status.AvailableReplicas) / float32(status.Replicas)
	if level <= *dh.critLevel {
		res.Status = CRITICAL
	} else if level <= *dh.warnLevel {
		res.Status = WARN
	}

	o, _ := json.MarshalIndent(status, "", "  ")
	res.Output = string(o)

	return res, nil
}

// register factory
func init() {
	RegisterCheck("deployment_status", NewDeploymentStatus, []string{"deployment"})
}