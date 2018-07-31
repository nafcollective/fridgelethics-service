package polling_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPolling(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Polling Suite")
}
