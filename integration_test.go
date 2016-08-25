package main_test

import (
	"os/exec"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("cfdot Integration", func() {
	Context("actual-lrp-groups", func() {
		var (
			sess *gexec.Session
		)

		Context("when no flags are passed", func() {

			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
						ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
							ActualLrpGroups: []*models.ActualLRPGroup{
								{
									Instance: &models.ActualLRP{
										State: "running",
									},
								},
							},
						}),
					),
				)
			})

			JustBeforeEach(func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "actual-lrp-groups")

				var err error
				sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
			})

			It("exits with status code of 0", func() {
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("returns the json encoding of the actual lrp", func() {
				Expect(sess.Out).To(gbytes.Say(`"state":"running"`))
			})
		})

		Context("when passing filters", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
						ghttp.VerifyProtoRepresenting(&models.ActualLRPGroupsRequest{
							Domain: "cf-apps",
							CellId: "cell_z1-0",
						}),
						ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
							ActualLrpGroups: []*models.ActualLRPGroup{
								{
									Instance: &models.ActualLRP{
										State: "running",
									},
								},
							},
						}),
					),
				)
			})

			It("exits with status code of 0", func() {
				cfdotCmd := exec.Command(
					cfdotPath,
					"--bbsURL", bbsServer.URL(),
					"actual-lrp-groups",
					"-d", "cf-apps",
					"-c", "cell_z1-0",
				)

				var err error
				sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("exits with status code of 0", func() {
				cfdotCmd := exec.Command(
					cfdotPath,
					"--bbsURL", bbsServer.URL(),
					"actual-lrp-groups",
					"--domain", "cf-apps",
					"--cell-id", "cell_z1-0",
				)

				var err error
				sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})

		})
	})

	Context("when the server responds with domains", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					ghttp.RespondWithProto(200, &models.DomainsResponse{
						Error:   nil,
						Domains: []string{"domain-1", "domain-2"},
					}),
				),
			)
		})

		It("domains prints a json stream of all the domains", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
		})
	})

	Context("when the server doesn't respond with domains", func() {
		BeforeEach(func() {
			bbsServer.RouteToHandler("POST", "/v1/domains/list",
				ghttp.RespondWith(500, []byte{}))
		})

		It("domains fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(4))

			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})

	Context("when the server returns an error", func() {
		var (
			sess *gexec.Session
		)

		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					ghttp.RespondWithProto(200, &models.DomainsResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "the request failed due to deadlock",
						},
						Domains: nil,
					}),
				),
			)
		})

		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
		})

		It("exits with status code 4", func() {
			Expect(sess.ExitCode()).To(Equal(4))
		})

		It("should print the type and message of the error", func() {
			Expect(sess.Err).To(gbytes.Say("BBS error"))
			Expect(sess.Err).To(gbytes.Say("Type 28: Deadlock"))
			Expect(sess.Err).To(gbytes.Say("Message: the request failed due to deadlock"))
		})

		It("should not print the usage", func() {
			Expect(sess.Err).NotTo(gbytes.Say("Usage:"))
		})
	})

	Describe("flag parsing for bbsURL", func() {
		Context("when running domains", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/domains/list"),
						ghttp.RespondWithProto(200, &models.DomainsResponse{}),
					),
				)
			})

			It("works with a --bbsURL flag specified before domains", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("works with a --bbsURL flag specified after domains", func() {
				cfdotCmd := exec.Command(cfdotPath, "domains", "--bbsURL", bbsServer.URL())

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})

		Context("when running set-domain", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/domains/upsert"),
						ghttp.RespondWithProto(200, &models.UpsertDomainResponse{}),
					),
				)
			})

			It("works with a --bbsURL flag specified before set-domain", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "anything", "--ttl", "0")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("works with a --bbsURL flag specified after set-domain", func() {
				cfdotCmd := exec.Command(cfdotPath, "set-domain", "--bbsURL", bbsServer.URL(), "anything", "--ttl", "40")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})
	})

	Context("when the server responds for set-domain", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/upsert"),
					ghttp.RespondWithProto(200, &models.UpsertDomainResponse{}),
				),
			)
		})

		It("set-domain works with a TTL not specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

		})

		It("set-domain works with a TTL specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "--ttl", "40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

		})

		It("set-domain prints to stderr when no domain specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "", "--ttl", "40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`No domain given`))
			Expect(sess.Err).To(gbytes.Say(`Usage`))
		})

		It("set-domain prints to stderr for negative TTL", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "--ttl", "-40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`ttl is negative`))
			Expect(sess.Err).To(gbytes.Say(`Usage:`))
		})

		It("set-domain prints to stderr for non-numeric TTL", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "-t", "asdf")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`ttl is non-numeric`))
			Expect(sess.Err).To(gbytes.Say(`Usage:`))
		})
	})

	Context("when the server does not respond for set-domain", func() {
		BeforeEach(func() {
			bbsServer.RouteToHandler("POST", "/v1/domains/upsert",
				ghttp.RespondWith(500, []byte{}))
		})
		It("set-domain fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(4))

			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})
})
