package taal_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kmacoskey/taal"
)

var _ = Describe("Taal", func() {

	var (
		t                              *Infra
		emptyInfra                     *Infra
		validTerraformConfig           []byte
		validTerraformConfigWithInputs []byte
		validTerraformCredentials      []byte
		validTerraformInputs           map[string]string
		validTerraformApplyStdout      string
		validTerraformDestroyStdout    string
		validPluginDir                 string
		emptyTerraformConfig           []byte
		emptyTerraformCredentials      []byte
		nilTerraformConfig             []byte
		nilTerraformCredentials        []byte
		invalidTerraformConfig         []byte
		invalidTerraformCredentials    []byte
		invalidTerraformError          string
		invalidTerraformStdout         string
		err                            error
		stdout                         string
	)

	BeforeEach(func() {
		emptyInfra = &Infra{}
		emptyTerraformConfig = []byte(``)
		emptyTerraformCredentials = []byte(``)
		nilTerraformConfig = []byte(``)
		nilTerraformCredentials = []byte(``)
		validTerraformApplyStdout = ApplySuccess
		validTerraformDestroyStdout = DestroySuccess

		validPluginDir, err = ioutil.TempDir("", "terraform_client_workingdir")
		Expect(err).NotTo(HaveOccurred())

		// TODO: Refactor this whole thing into a set of functions
		terraformGooglePluginUrl := "https://releases.hashicorp.com/terraform-provider-google/1.16.2/terraform-provider-google_1.16.2_darwin_amd64.zip"
		terraformGooglePluginFilePath := filepath.Join(validPluginDir, "terraform-provider-google_1.16.2_darwin_amd64.zip")
		err := downloadFile(terraformGooglePluginFilePath, terraformGooglePluginUrl)
		Expect(err).NotTo(HaveOccurred())
		cmdEnv := []string{
			fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		}
		args := []string{terraformGooglePluginFilePath, fmt.Sprintf("-d %s", validPluginDir)}
		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("unzip %s", strings.Join(args, " ")))
		cmd.Env = cmdEnv
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())

		validTerraformConfig = []byte(`provider "google" { project = "data-gp-toolsmiths" region = "us-central1" } resource "google_compute_project_metadata_item" "default" { key = "my_metadata" value = "my_value" }`)
		validTerraformConfigWithInputs = []byte(`variable "key" {} provider "google" { project = "data-gp-toolsmiths" region = "us-central1" } resource "google_compute_project_metadata_item" "default" { key = "my_metadata" value = "my_value" }`)
		validTerraformCredentials = []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
		validTerraformInputs = map[string]string{"key": "value"}
		invalidTerraformConfig = []byte(`foo`)
		invalidTerraformCredentials = []byte(`foo`)
		invalidTerraformError = "foo"
		invalidTerraformStdout = "foo"
	})

	Describe("Creating new infrastructure", func() {

		Context("When requesting a new infra object", func() {
			It("Should return an infra object", func() {
				infra := NewInfra()
				Expect(infra).To(Equal(emptyInfra))
			})
		})

	})

	Describe("Setting infrastructure values", func() {
		BeforeEach(func() {
			t = NewInfra()
		})

		Context("When setting the config", func() {
			It("Should not error", func() {
				t.Config([]byte(`foo`))
				Expect(t.GetConfig()).To(Equal([]byte(`foo`)))
			})
		})

		Context("When setting the credentials", func() {
			It("Should not error", func() {
				t.Credentials([]byte(`foo`))
				Expect(t.GetCredentials()).To(Equal([]byte(`foo`)))
			})
		})

		Context("When setting the plugin directory", func() {
			It("Should not error", func() {
				t.PluginDir(validPluginDir)
				Expect(t.GetPluginDir()).To(Equal(validPluginDir))
			})
		})

	})

	// ###########################################
	//                    _
	//   __ _ _ __  _ __ | |_   _
	//  / _` | '_ \| '_ \| | | | |
	// | (_| | |_) | |_) | | |_| |
	//  \__,_| .__/| .__/|_|\__, |
	//       |_|   |_|      |___/
	//
	// ###########################################

	Describe("Applying infrastructure", func() {
		BeforeEach(func() {
			t = NewInfra()
			t.PluginDir(validPluginDir)
		})

		Context("When everything goes ok", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfig)
				t.Credentials(validTerraformCredentials)
				stdout, err = t.Apply()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout without errors", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(validTerraformApplyStdout))
			})
		})

		Context("When user inputs are specified", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfigWithInputs)
				t.Credentials(validTerraformCredentials)
				t.Inputs(validTerraformInputs)
				stdout, err = t.Apply()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout without errors", func() {
				Expect(stdout).To(ContainSubstring(validTerraformApplyStdout))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When there is no terraform configuration set", func() {
			It("Should error when configuration is empty", func() {
				t.Config(emptyTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, err := t.Apply()
				Expect(err).To(MatchError(ErrorMissingConfig))
			})
			It("Should error when configuration is nil", func() {
				t.Config(nilTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, err := t.Apply()
				Expect(err).To(MatchError(ErrorMissingConfig))
			})
		})

		Context("When there are no credentials set", func() {
			It("Should error when credentials is empty", func() {
				t.Config(validTerraformConfig)
				t.Credentials(emptyTerraformCredentials)
				_, err := t.Apply()
				Expect(err).To(MatchError(ErrorMissingCredentials))
			})
			It("Should error when credentials is nil", func() {
				t.Config(validTerraformConfig)
				t.Credentials(nilTerraformCredentials)
				_, err := t.Apply()
				Expect(err).To(MatchError(ErrorMissingCredentials))
			})
		})

		Context("When the terraform apply fails", func() {
			BeforeEach(func() {
				t.Config(invalidTerraformConfig)
				t.Credentials(validTerraformCredentials)
				stdout, err = t.Apply()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout and error", func() {
				Expect(stdout).To(ContainSubstring(PlanFailure))
				Expect(err).To(HaveOccurred())
			})
		})

	})

	// ###########################################
	//      _           _
	//   __| | ___  ___| |_ _ __ ___  _   _
	//  / _` |/ _ \/ __| __| '__/ _ \| | | |
	// | (_| |  __/\__ \ |_| | | (_) | |_| |
	//  \__,_|\___||___/\__|_|  \___/ \__, |
	//                                |___/
	//
	// ###########################################

	Describe("Destroying infrastructure", func() {
		BeforeEach(func() {
			t = NewInfra()
			t.PluginDir(validPluginDir)
		})

		Context("When everything goes ok", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, applyErr := t.Apply()
				Expect(applyErr).NotTo(HaveOccurred())
				stdout, err = t.Destroy()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout and not error", func() {
				Expect(stdout).To(ContainSubstring(validTerraformDestroyStdout))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When user inputs are provided", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfigWithInputs)
				t.Credentials(validTerraformCredentials)
				t.Inputs(validTerraformInputs)
				_, applyErr := t.Apply()
				Expect(applyErr).NotTo(HaveOccurred())
				stdout, err = t.Destroy()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout and not error", func() {
				Expect(stdout).To(ContainSubstring(validTerraformDestroyStdout))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When there is no terraform configuration set", func() {
			It("Should error when configuration is empty", func() {
				t.Config(emptyTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, err = t.Destroy()
				Expect(err).To(MatchError(ErrorMissingConfig))
			})
			It("Should error when configuration is nil", func() {
				t.Config(nilTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, err = t.Destroy()
				Expect(err).To(MatchError(ErrorMissingConfig))
			})
		})

		Context("When there are no credentials set", func() {
			It("Should error when credentials is empty", func() {
				t.Config(validTerraformConfig)
				t.Credentials(emptyTerraformCredentials)
				_, err := t.Destroy()
				Expect(err).To(MatchError(ErrorMissingCredentials))
			})
			It("Should error when credentials is nil", func() {
				t.Config(validTerraformConfig)
				t.Credentials(nilTerraformCredentials)
				_, err := t.Destroy()
				Expect(err).To(MatchError(ErrorMissingCredentials))
			})
		})

		Context("When the terraform destroy fails", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, applyErr := t.Apply()
				Expect(applyErr).NotTo(HaveOccurred())
				t.Config(invalidTerraformConfig)
				stdout, err = t.Destroy()
			})
			// Overloaded to avoid execessive testing time
			It("Should return the expected stdout and error", func() {
				Expect(err).To(HaveOccurred())
				Expect(stdout).To(ContainSubstring(PlanFailure))
			})
		})

	})

})

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
