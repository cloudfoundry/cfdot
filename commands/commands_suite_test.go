package commands_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var _ = SynchronizedBeforeSuite(func() []byte {
	chmodErr := os.Chmod("fixtures/bbsClientBadPermissions.key", 0300)
	Expect(chmodErr).NotTo(HaveOccurred())

	return nil
}, func(_ []byte) {})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	chmodErr := os.Chmod("fixtures/bbsClientBadPermissions.key", 0644)
	Expect(chmodErr).NotTo(HaveOccurred())
})

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}
