package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("claim-lock", func() {
	itValidatesLocketFlags("claim-lock")

	var (
		locketClientCertFile string
		locketClientKeyFile  string
		logger               *lagertest.TestLogger
	)

	BeforeEach(func() {
		wd, _ := os.Getwd()
		locketClientCertFile = fmt.Sprintf("%s/fixtures/locketClient.crt", wd)
		locketClientKeyFile = fmt.Sprintf("%s/fixtures/locketClient.key", wd)
		os.Setenv("LOCKET_CA_CERT_FILE", locketCACertFile)
		os.Setenv("LOCKET_CERT_FILE", locketClientCertFile)
		os.Setenv("LOCKET_KEY_FILE", locketClientKeyFile)

		logger = lagertest.NewTestLogger("locket")
	})

	It("claim-lock exists successfully with code 0", func() {
		cfdotCmd := exec.Command(cfdotPath,
			"--locketAPILocation", locketAPILocation,
			"claim-lock",
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
				"claim-lock",
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
