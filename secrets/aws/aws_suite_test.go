package aws_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAwsSecrets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "aws secret Suite")
}
