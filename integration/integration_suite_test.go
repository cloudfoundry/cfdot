package integration_test

import (
	"encoding/json"
	"os"
	"os/exec"

	"code.cloudfoundry.org/bbs/test_helpers"
	"code.cloudfoundry.org/bbs/test_helpers/sqlrunner"
	"code.cloudfoundry.org/consuladapter/consulrunner"
	"code.cloudfoundry.org/locket/cmd/locket/config"
	"code.cloudfoundry.org/locket/cmd/locket/testrunner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"fmt"
	"testing"
)

var (
	cfdotPath, locketPath string
	locketRunner          *ginkgomon.Runner
	locketProcess         ifrit.Process
	dbRunner              sqlrunner.SQLRunner
	dbProcess             ifrit.Process
	consulRunner          *consulrunner.ClusterRunner
	locketAPILocation     string
	locketCACertFile      string
)

var bbsServer *ghttp.Server

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("code.cloudfoundry.org/cfdot")
	Expect(err).NotTo(HaveOccurred())

	locketBinPath, err := gexec.Build("code.cloudfoundry.org/locket/cmd/locket")
	Expect(err).NotTo(HaveOccurred())

	bytes, err := json.Marshal([]string{binPath, locketBinPath})
	Expect(err).NotTo(HaveOccurred())

	return []byte(bytes)
}, func(data []byte) {
	paths := []string{}
	err := json.Unmarshal(data, &paths)
	Expect(err).NotTo(HaveOccurred())
	cfdotPath = paths[0]
	locketPath = paths[1]
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	bbsServer = ghttp.NewServer()
	port := 8090 + GinkgoParallelNode()
	locketAPILocation = fmt.Sprintf("localhost:%d", port)
	wd, _ := os.Getwd()
	locketCACertFile = fmt.Sprintf("%s/fixtures/locketCA.crt", wd)

	dbName := fmt.Sprintf("diego_%d", GinkgoParallelNode())
	dbRunner = test_helpers.NewSQLRunner(dbName)
	dbProcess = ginkgomon.Invoke(dbRunner)

	consulRunner = consulrunner.NewClusterRunner(
		consulrunner.ClusterRunnerConfig{
			StartingPort: 9001 + GinkgoParallelNode()*consulrunner.PortOffsetLength,
			NumNodes:     1,
			Scheme:       "http",
		},
	)
	consulRunner.Start()

	locketRunner = testrunner.NewLocketRunner(locketPath, func(cfg *config.LocketConfig) {
		cfg.CaFile = locketCACertFile
		cfg.ListenAddress = locketAPILocation
		cfg.ConsulCluster = consulRunner.ConsulCluster()
		cfg.DatabaseDriver = dbRunner.DriverName()
		cfg.DatabaseConnectionString = dbRunner.ConnectionString()
	})
	locketProcess = ginkgomon.Invoke(locketRunner)
})

var _ = AfterEach(func() {
	bbsServer.CloseClientConnections()
	bbsServer.Close()
	ginkgomon.Interrupt(locketProcess)
	dbProcess.Signal(os.Interrupt)
	consulRunner.Stop()
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

// Pass arguments that would be passed to cfdot
func itValidatesLocketFlags(args ...string) {
	Context("Locket Flag Validation", func() {
		It("exits with status 3 when no locket flags are specified", func() {
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
			urlFlag := "--bbsURL"
			url := bbsServer.URL()
			if command == "locks" {
				urlFlag = "--locketAPILocation"
				url = locketAPILocation
			}
			cfdotCmd := exec.Command(cfdotPath, urlFlag, url, command, "extra-arg")

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
