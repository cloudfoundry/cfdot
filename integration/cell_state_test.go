package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/rep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("cell-state", func() {
	itValidatesTLSFlags("cell-state")

	Context("when cell-state command is called", func() {
		var presence1, presence2 *models.CellPresence
		var rep1Server, rep2Server *ghttp.Server
		var cellState1, cellState2 *rep.CellState
		var (
			clientCAFile   string
			clientCertFile string
			clientKeyFile  string
		)

		BeforeEach(func() {
			rep1Server = ghttp.NewUnstartedServer()
			wd, _ := os.Getwd()
			clientCAFile = fmt.Sprintf("%s/fixtures/locketCA.crt", wd)
			clientCertFile = fmt.Sprintf("%s/fixtures/locketClient.crt", wd)
			clientKeyFile = fmt.Sprintf("%s/fixtures/locketClient.key", wd)

			tlsConfig, err := cfhttp.NewTLSConfig(clientCertFile, clientKeyFile, clientCAFile)
			Expect(err).NotTo(HaveOccurred())
			rep1Server.HTTPTestServer.TLS = tlsConfig
			rep1Server.HTTPTestServer.StartTLS()

			rep2Server = ghttp.NewServer()

			presence1 = &models.CellPresence{
				CellId: "cell-1",
				RepUrl: rep1Server.URL(),
			}
			presence2 = &models.CellPresence{
				CellId: "cell-2",
				RepUrl: rep2Server.URL(),
			}

			cellState1 = &rep.CellState{
				RepURL:             rep1Server.URL(),
				CellID:             "cell-1",
				RootFSProviders:    rep.RootFSProviders{},
				AvailableResources: rep.Resources{},
				TotalResources:     rep.Resources{},
				LRPs:               []rep.LRP{},
				Tasks:              []rep.Task{},
			}
			cellState2 = &rep.CellState{
				RepURL:             rep2Server.URL(),
				CellID:             "cell-2",
				RootFSProviders:    rep.RootFSProviders{},
				AvailableResources: rep.Resources{},
				TotalResources:     rep.Resources{},
				LRPs:               []rep.LRP{},
				Tasks:              []rep.Task{},
			}

			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
					ghttp.RespondWithProto(200, &models.CellsResponse{
						Cells: []*models.CellPresence{presence1, presence2},
					}),
				),
			)
			rep1Server.RouteToHandler("GET", "/state", func(resp http.ResponseWriter, req *http.Request) {
				jsonData, err := json.Marshal(cellState1)
				Expect(err).NotTo(HaveOccurred())
				resp.Write(jsonData)
			})

			rep2Server.RouteToHandler("GET", "/state", func(resp http.ResponseWriter, req *http.Request) {
				jsonData, err := json.Marshal(cellState2)
				Expect(err).NotTo(HaveOccurred())
				resp.Write(jsonData)
			})
		})

		AfterEach(func() {
			rep1Server.CloseClientConnections()
			rep1Server.Close()
			rep2Server.CloseClientConnections()
			rep2Server.Close()
		})

		It("returns the json encoding of the correct cell-state", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-state", "cell-2")
			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))

			jsonData, err := json.Marshal(cellState2)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes.TrimSpace(sess.Out.Contents())).To(Equal(jsonData))
		})

		Context("when timeout flag is present", func() {
			var (
				serverTimeout int
				sess          *gexec.Session
			)

			JustBeforeEach(func() {
				bbsServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
						func(w http.ResponseWriter, req *http.Request) {
							time.Sleep(time.Duration(serverTimeout) * time.Second)
						},
						ghttp.RespondWithProto(200, &models.CellsResponse{
							Cells: []*models.CellPresence{presence1, presence2},
						}),
					),
				)

				cfdotCmd := exec.Command(
					cfdotPath,
					"--bbsURL", bbsServer.URL(),
					"--timeout", "1",
					"cell-state", "cell-2",
				)
				var err error
				sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when exceeds timeout", func() {
				BeforeEach(func() {
					serverTimeout = 2
				})

				It("exits with code 4 and a timeout message", func() {
					Eventually(sess, 2).Should(gexec.Exit(4))
					Expect(sess.Err).To(gbytes.Say(`Timeout exceeded`))
				})

			})

			Context("when within timeout", func() {
				BeforeEach(func() {
					serverTimeout = 0
				})

				It("should succeed", func() {
					Eventually(sess).Should(gexec.Exit(0))
					jsonData, err := json.Marshal(cellState2)
					Expect(err).NotTo(HaveOccurred())
					Expect(bytes.TrimSpace(sess.Out.Contents())).To(Equal(jsonData))
				})
			})
		})

		Context("when the rep has mutual TLS enabled", func() {
			var args []string

			It("uses the correct TLS config", func() {
				args = []string{
					"--bbsURL", bbsServer.URL(),
					"--caCertFile", clientCAFile,
					"--clientCertFile", clientCertFile,
					"--clientKeyFile", clientKeyFile,
					"cell-state", "cell-1",
				}
				cfdotCmd := exec.Command(cfdotPath, args...)
				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))

				jsonData, err := json.Marshal(cellState1)
				Expect(err).NotTo(HaveOccurred())
				Expect(bytes.TrimSpace(sess.Out.Contents())).To(Equal(jsonData))
			})

			Context("cell-states", func() {
				It("returns the json encoding of the cell-states", func() {
					args = []string{
						"--bbsURL", bbsServer.URL(),
						"--caCertFile", clientCAFile,
						"--clientCertFile", clientCertFile,
						"--clientKeyFile", clientKeyFile,
						"cell-states",
					}
					cfdotCmd := exec.Command(cfdotPath, args...)
					sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))

					decoder := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(sess.Out.Contents())))
					var receivedState rep.CellState

					err = decoder.Decode(&receivedState)
					Expect(err).NotTo(HaveOccurred())
					Expect(receivedState).To(Equal(*cellState1))

					err = decoder.Decode(&receivedState)
					Expect(err).NotTo(HaveOccurred())
					Expect(receivedState).To(Equal(*cellState2))
				})
			})

			Context("when timeout flag is present", func() {
				var (
					serverTimeout int
					sess          *gexec.Session
				)

				JustBeforeEach(func() {
					bbsServer.SetHandler(0,
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
							func(w http.ResponseWriter, req *http.Request) {
								time.Sleep(time.Duration(serverTimeout) * time.Second)
							},
							ghttp.RespondWithProto(200, &models.CellsResponse{
								Cells: []*models.CellPresence{presence1, presence2},
							}),
						),
					)

					cfdotCmd := exec.Command(
						cfdotPath,
						"--bbsURL", bbsServer.URL(),
						"--caCertFile", clientCAFile,
						"--clientCertFile", clientCertFile,
						"--clientKeyFile", clientKeyFile,
						"--timeout", "1",
						"cell-states",
					)
					var err error
					sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
				})

				Context("when exceeds timeout", func() {
					BeforeEach(func() {
						serverTimeout = 2
					})

					It("exits with code 4 and a timeout message", func() {
						Eventually(sess, 2).Should(gexec.Exit(4))
						Expect(sess.Err).To(gbytes.Say(`Timeout exceeded`))
					})
				})

				Context("when within timeout", func() {
					BeforeEach(func() {
						serverTimeout = 0
					})

					It("should succeed", func() {
						Eventually(sess).Should(gexec.Exit(0))
						decoder := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(sess.Out.Contents())))
						var receivedState rep.CellState

						err := decoder.Decode(&receivedState)
						Expect(err).NotTo(HaveOccurred())
						Expect(receivedState).To(Equal(*cellState1))

						err = decoder.Decode(&receivedState)
						Expect(err).NotTo(HaveOccurred())
						Expect(receivedState).To(Equal(*cellState2))
					})
				})
			})
		})

		Context("when the cell does not exist", func() {
			It("exits with status code of 5", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-state", "cell-id-dsafasdklfjasdlkf")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(5))
			})
		})

		Context("when the BBS request fails", func() {
			BeforeEach(func() {
				bbsServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
						ghttp.RespondWithProto(503, &models.CellsResponse{
							Error: models.ErrUnknownError,
						}),
					),
				)
			})

			It("exits with status code of 4", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-state", "cell-2")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(4))
				Expect(sess.Err).To(gbytes.Say("BBS error"))
			})

			Context("cell-states", func() {
				It("exits with status code of 4", func() {
					cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-states")

					sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(4))
					Expect(sess.Err).To(gbytes.Say("BBS error"))
					Expect(sess.Err).To(gbytes.Say("Failed to get cell registrations from BBS"))
				})
			})
		})

		Context("when the Rep request fails", func() {
			BeforeEach(func() {
				rep2Server.RouteToHandler("GET", "/state", func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(503)
				})
			})

			It("exits with status code of 4", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-state", "cell-2")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(4))
				Expect(sess.Err).To(gbytes.Say("Rep error"))
			})

			Context("cell-states", func() {
				It("exits with status code of 4", func() {
					cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-states")

					sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(4))
					Expect(sess.Err).To(gbytes.Say("Rep error"))
					Expect(sess.Err).To(gbytes.Say("Failed to get cell state for cell cell-1"))
					Expect(sess.Err).To(gbytes.Say("Failed to get cell state for cell cell-2"))
				})
			})
		})

		Context("when cell command is called with extra arguments", func() {
			It("exits with status code of 3", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-state", "cell-id", "extra-argument")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(3))
			})

			Context("cell-states", func() {
				It("exits with status code of 3", func() {
					cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell-states", "extra-argument")

					sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(3))
				})
			})
		})
	})
})
