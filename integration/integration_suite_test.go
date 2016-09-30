package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"fmt"
	"testing"
)

var cfdotPath string

var bbsServer *ghttp.Server

const targetName = "testserver"

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("code.cloudfoundry.org/cfdot")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	cfdotPath = string(data)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	bbsServer = ghttp.NewServer()
})

var _ = AfterEach(func() {
	bbsServer.Close()
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

// Pass arguments that would be passed to cfdot
// i.e. set-domain domain1
func itValidatesBBSFlags(args ...string) {
	Context("BBS Flag Validation", func() {
		It("exits with status 3 when no bbs flags are specified", func() {
			cmd := exec.Command(cfdotPath, args...)

			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Exited).Should(BeClosed())

			Expect(sess.ExitCode()).To(Equal(3))
		})
	})
}

func itHasNoArgs(command string) {
	var (
		sess *gexec.Session
	)
	Context("when any arguments are passed", func() {
		BeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), command, "extra-arg")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status code of 3", func() {
			Expect(sess.ExitCode()).To(Equal(3))
		})

		It("prints the usage to stderr", func() {
			Expect(sess.Err).To(gbytes.Say(fmt.Sprintf("cfdot %s \\[flags\\]", command)))
		})
	})
}
