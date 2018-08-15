package taal

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ErrorMissingCredentials = "no credentials specified for running terraform actions"
	ErrorMissingConfig      = "no configuration supplied for running terraform actions"
	ApplySuccess            = "Apply complete! Resources:"
	DestroySuccess          = "Destroy complete! Resources:"
	PlanFailure             = "There are some problems with the configuration"
)

type Terraform interface {
	Config([]byte)
	GetConfig() []byte
	Credentials([]byte)
	GetCredentials() []byte
	Apply() (string, error)
	Destroy() (string, error)
}

type Infra struct {
	config      []byte
	credentials []byte
	state       []byte
	pluginDir   string
}

func NewInfra() *Infra {
	return &Infra{}
}

func (terraform *Infra) Config(config []byte) {
	terraform.config = config
}

func (terraform *Infra) GetConfig() []byte {
	return terraform.config
}

func (terraform *Infra) Credentials(credentials []byte) {
	terraform.credentials = credentials
}

func (terraform *Infra) GetCredentials() []byte {
	return terraform.credentials
}

func (terraform *Infra) State(state []byte) {
	terraform.state = state
}

func (terraform *Infra) GetState() []byte {
	return terraform.state
}

func (terraform *Infra) PluginDir(pluginDir string) {
	terraform.pluginDir = pluginDir
}

func (terraform *Infra) GetPluginDir() string {
	return terraform.pluginDir
}

func (terraform *Infra) Apply() (string, error) {

	credentials := terraform.GetCredentials()
	config := terraform.GetConfig()
	pluginDir := terraform.GetPluginDir()

	if len(credentials) == 0 {
		return "", errors.New(ErrorMissingCredentials)
	}

	if len(config) == 0 {
		return "", errors.New(ErrorMissingConfig)
	}

	wd, err := ioutil.TempDir("", "terraform_client_workingdir")
	if err != nil {
		return "", err
	}

	configfilename := "terraform.tf"
	statefilename := "terraform.tfstate"

	configfile := filepath.Join(wd, configfilename)
	err = ioutil.WriteFile(configfile, config, 0666)
	if err != nil {
		return "", err
	}

	initArgs := []string{
		"init",
		"-input=false",
		"-get=true",
		"-backend=false",
	}

	if len(pluginDir) > 0 {
		initArgs = append(initArgs, fmt.Sprintf("-plugin-dir=%s", pluginDir))
	}

	initArgs = append(initArgs, wd)

	cmdEnv := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", credentials),
		// https://github.com/hashicorp/terraform/blob/master/vendor/github.com/hashicorp/go-checkpoint/README.md
		"CHECKPOINT_DISABLE=1",
		// "TF_LOG=DEBUG",
	}

	err, stdout, stderr := run(wd, cmdEnv, initArgs)
	if err != nil {
		return stderr, err
	}

	applyArgs := []string{
		"apply",
		"-auto-approve",
		"-input=false", // do not prompt for inputs
	}

	err, stdout, stderr = run(wd, cmdEnv, applyArgs)
	if err != nil {
		return stderr, err
	}

	statefile := filepath.Join(wd, statefilename)
	state, err := ioutil.ReadFile(statefile)
	if err != nil {
		return "", err
	}

	terraform.State(state)

	return stdout, nil
}

func (terraform *Infra) Destroy() (string, error) {
	credentials := terraform.GetCredentials()
	config := terraform.GetConfig()
	state := terraform.GetState()
	pluginDir := terraform.GetPluginDir()

	if len(credentials) == 0 {
		return "", errors.New(ErrorMissingCredentials)
	}

	if len(config) == 0 {
		return "", errors.New(ErrorMissingConfig)
	}

	wd, err := ioutil.TempDir("", "terraform_client_workingdir")
	if err != nil {
		return "", err
	}

	configfilename := "terraform.tf"
	statefilename := "terraform.tfstate"

	configfile := filepath.Join(wd, configfilename)
	err = ioutil.WriteFile(configfile, config, 0666)
	if err != nil {
		return "", err
	}

	statefile := filepath.Join(wd, statefilename)
	err = ioutil.WriteFile(statefile, state, 0666)
	if err != nil {
		return "", err
	}

	initArgs := []string{
		"init",
		"-input=false",
		"-get=true",
		"-backend=false",
	}

	if len(pluginDir) > 0 {
		initArgs = append(initArgs, fmt.Sprintf("-plugin-dir=%s", pluginDir))
	}

	initArgs = append(initArgs, wd)

	cmdEnv := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", credentials),
		// https://github.com/hashicorp/terraform/blob/master/vendor/github.com/hashicorp/go-checkpoint/README.md
		"CHECKPOINT_DISABLE=1",
		// "TF_LOG=DEBUG",
	}

	err, stdout, stderr := run(wd, cmdEnv, initArgs)
	if err != nil {
		return stderr, err
	}

	destroyArgs := []string{
		"destroy",
		"-force",
	}

	err, stdout, stderr = run(wd, cmdEnv, destroyArgs)
	if err != nil {
		return stderr, err
	}

	return stdout, nil
}

func run(directory string, environment []string, args []string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	defaultArgs := []string{
		"-no-color",
	}

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("terraform %s %s", strings.Join(args, " "), strings.Join(defaultArgs, " ")))
	cmd.Env = environment

	// When no directory is specified, a location to run terraform is still required
	//  therefore a temporary directory is used.
	if len(directory) == 0 {
		temp_work_dir, err := ioutil.TempDir("", "terraform_client_workingdir")
		if err != nil {
			return err, "", ""
		}
		cmd.Dir = temp_work_dir
	} else {
		cmd.Dir = directory
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return err, stdout.String(), stderr.String()
}
