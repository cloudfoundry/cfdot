package integration_test

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("claim-presence", func() {
	itValidatesLocketFlags("claim-presence")

	BeforeEach(func() {
		os.Setenv("CA_CERT_FILE", locketCACertFile)
		os.Setenv("CLIENT_CERT_FILE", locketClientCertFile)
		os.Setenv("CLIENT_KEY_FILE", locketClientKeyFile)
	})

	AfterEach(func() {
		os.Unsetenv("CA_CERT_FILE")
		os.Unsetenv("CLIENT_CERT_FILE")
		os.Unsetenv("CLIENT_KEY_FILE")
	})

	It("claim-lock exists successfully with code 0", func() {
		cfdotCmd := exec.Command(cfdotPath,
			"--locketAPILocation", locketAPILocation,
			"claim-presence",
			"--key", "test-key",
			"--owner", "test-owner",
			"--ttl", "60",
		)

		sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(sess).Should(gexec.Exit(0))
	})

	Context("when the server is down", func() {
		BeforeEach(func() {
			ginkgomon.Interrupt(locketProcess)
		})

		It("claim-lock fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath,
				"--locketAPILocation", locketAPILocation,
				"claim-presence",
				"--key", "test-key",
				"--owner", "test-owner",
				"--ttl", "60",
			)

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess, 2*time.Second).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("connection refused"))
		})
	})
})
