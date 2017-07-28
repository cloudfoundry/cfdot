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

var _ = Describe("Locket commands", func() {
	var (
		locketClientCertFile string
		locketClientKeyFile  string
		logger               *lagertest.TestLogger
		locketClient         models.LocketClient
	)

	BeforeEach(func() {
		wd, _ := os.Getwd()
		locketClientCertFile = fmt.Sprintf("%s/fixtures/locketClient.crt", wd)
		locketClientKeyFile = fmt.Sprintf("%s/fixtures/locketClient.key", wd)
		os.Setenv("LOCKET_CA_CERT_FILE", locketCACertFile)
		os.Setenv("LOCKET_CERT_FILE", locketClientCertFile)
		os.Setenv("LOCKET_KEY_FILE", locketClientKeyFile)

		var err error
		locketClient, err = locket.NewClient(logger, locket.ClientLocketConfig{
			LocketAddress:        locketAPILocation,
			LocketCACertFile:     locketCACertFile,
			LocketClientCertFile: locketClientCertFile,
			LocketClientKeyFile:  locketClientKeyFile,
		})
		Expect(err).NotTo(HaveOccurred())

		logger = lagertest.NewTestLogger("locket")
	})

	Describe("release-lock", func() {
		itValidatesLocketFlags("release-lock")

		BeforeEach(func() {
			req := &models.LockRequest{
				Resource: &models.Resource{
					Key:   "test-key",
					Owner: "test-owner",
					Value: "test-value",
					Type:  "lock",
				},
				TtlInSeconds: 10,
			}

			_, err := locketClient.Lock(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits successfully with code 0 after releasing the lock", func() {
			cfdotCmd := exec.Command(cfdotPath,
				"--locketAPILocation", locketAPILocation,
				"release-lock",
				"--key", "test-key",
				"--owner", "test-owner",
			)

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))

			_, err = locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test-key"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("claim-locks", func() {
		itValidatesLocketFlags("claim-lock")

		It("exits successfully with code 0 after acquiring the lock", func() {
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

			resp, err := locketClient.Fetch(context.Background(), &models.FetchRequest{Key: "test-key"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetResource()).To(Equal(&models.Resource{
				Key:      "test-key",
				Owner:    "test-owner",
				Type:     "lock",
				TypeCode: models.LOCK,
			}))
		})

		Context("when the server is down", func() {
			BeforeEach(func() {
				ginkgomon.Interrupt(locketProcess)
			})

			It("fails with a relevant error message", func() {
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

	Describe("locks", func() {
		itValidatesLocketFlags("locks")
		itHasNoArgs("locks", true)

		Context("when the server responds with locks", func() {
			BeforeEach(func() {
				req := &models.LockRequest{
					Resource: &models.Resource{
						Key:   "key",
						Owner: "owner",
						Value: "value",
						Type:  "lock",
					},
					TtlInSeconds: 10,
				}

				_, err := locketClient.Lock(context.Background(), req)
				Expect(err).NotTo(HaveOccurred())
			})

			It("prints a json stream of all the locks", func() {
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

			It("fails with a relevant error message", func() {
				cfdotCmd := exec.Command(cfdotPath, "--locketAPILocation", locketAPILocation, "locks")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess, 2*time.Second).Should(gexec.Exit(4))
				Expect(sess.Err).To(gbytes.Say("connection refused"))
			})
		})
	})
})
