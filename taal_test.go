package taal_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kmacoskey/taal"
)

var _ = Describe("Taal", func() {

	var (
		emptyInfra                  *Infra
		t                           *Infra
		validTerraformConfig        []byte
		validTerraformCredentials   []byte
		validTerraformApplyStdout   string
		validTerraformDestroyStdout string
		emptyTerraformConfig        []byte
		emptyTerraformCredentials   []byte
		nilTerraformConfig          []byte
		nilTerraformCredentials     []byte
		invalidTerraformConfig      []byte
		invalidTerraformCredentials []byte
		invalidTerraformError       string
		invalidTerraformStdout      string
		err                         error
		stdout                      string
	)

	BeforeEach(func() {
		emptyInfra = &Infra{}
		emptyTerraformConfig = []byte(``)
		emptyTerraformCredentials = []byte(``)
		nilTerraformConfig = []byte(``)
		nilTerraformCredentials = []byte(``)
		validTerraformApplyStdout = ApplySuccess
		validTerraformDestroyStdout = DestroySuccess

		validTerraformConfig = []byte(`provider "google" { project = "data-gp-toolsmiths" region = "us-central1" } resource "google_compute_project_metadata_item" "default" { key = "my_metadata" value = "my_value" }`)
		validTerraformCredentials = []byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
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
		})

		Context("When everything goes ok", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfig)
				t.Credentials(validTerraformCredentials)
				stdout, err = t.Apply()
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return the expected stdout", func() {
				Expect(stdout).To(ContainSubstring(validTerraformApplyStdout))
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
			It("Should error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Should return the expected stdout", func() {
				Expect(stdout).To(ContainSubstring(PlanFailure))
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
		})

		Context("When everything goes ok", func() {
			BeforeEach(func() {
				t.Config(validTerraformConfig)
				t.Credentials(validTerraformCredentials)
				_, applyErr := t.Apply()
				Expect(applyErr).NotTo(HaveOccurred())
				stdout, err = t.Destroy()
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return the expected stdout", func() {
				Expect(stdout).To(ContainSubstring(validTerraformDestroyStdout))
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
			It("Should return the expected error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Should return the expected stdout", func() {
				Expect(stdout).To(ContainSubstring(PlanFailure))
			})
		})

	})

})
