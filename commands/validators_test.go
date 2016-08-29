package commands_test

import (
	"code.cloudfoundry.org/cfdot/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Validators", func() {

	var dummyCmd *cobra.Command

	BeforeEach(func() {
		dummyCmd = &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {},
		}
	})

	Context("Integer flag validation", func() {
		It("returns integer for valid input", func() {
			intValue, err := commands.ValidatePositiveIntegerForFlag("test-flag", "2", dummyCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(intValue).To(Equal(2))
		})

		It("returns error for invalid input", func() {
			_, err := commands.ValidatePositiveIntegerForFlag("test-flag", "incorrect-input", dummyCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test-flag is non-numeric"))
		})

		It("returns error for negative integer input", func() {
			_, err := commands.ValidatePositiveIntegerForFlag("test-flag", "-1", dummyCmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test-flag is negative"))
		})
	})

})
