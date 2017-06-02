package integration_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/locket"
	"code.cloudfoundry.org/locket/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("locks", func() {
	itValidatesLocketFlags("locks")
	itHasNoArgs("locks")

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

	Context("when the server responds with locks", func() {
		BeforeEach(func() {
			locketConfig := locket.ClientLocketConfig{
				LocketAddress:        locketAPILocation,
				LocketCACertFile:     locketCACertFile,
				LocketClientCertFile: locketClientCertFile,
				LocketClientKeyFile:  locketClientKeyFile,
			}
			locketClient, err := locket.NewClient(logger, locketConfig)
			Expect(err).NotTo(HaveOccurred())

			req := &models.LockRequest{
				Resource: &models.Resource{
					Key:   "key",
					Owner: "owner",
					Value: "value",
					Type:  "lock",
				},
				TtlInSeconds: 10,
			}

			_, err = locketClient.Lock(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("locks prints a json stream of all the locks", func() {
			cfdotCmd := exec.Command(cfdotPath, "--locketAPILocation", locketAPILocation, "locks")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say(`"key":"key","owner":"owner"`))
		})
	})

	Context("when the server is down", func() {
		BeforeEach(func() {
			ginkgomon.Interrupt(locketProcess)
		})

		It("locks fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--locketAPILocation", locketAPILocation, "locks")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess, 2*time.Second).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("the connection is unavailable"))
		})
	})
})
