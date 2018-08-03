package taal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTaal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Taal Suite")
}
