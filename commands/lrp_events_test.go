package commands_test

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/cfdot/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LRP Events", func() {
	var (
		fakeBBSClient           *fake_bbs.FakeClient
		fakeEventSource         *eventfakes.FakeEventSource
		fakeInstanceEventSource *eventfakes.FakeEventSource
		stdout, stderr          *gbytes.Buffer
		lrp                     *models.DesiredLRP
		actualLRP               *models.ActualLRP
	)

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
		fakeEventSource = &eventfakes.FakeEventSource{}
		fakeInstanceEventSource = &eventfakes.FakeEventSource{}
		fakeBBSClient.SubscribeToEventsByCellIDReturns(fakeEventSource, nil)
		fakeBBSClient.SubscribeToInstanceEventsByCellIDReturns(fakeInstanceEventSource, nil)

		lrp = &models.DesiredLRP{
			ProcessGuid: "some-desired-lrp",
		}

		actualLRP = model_helpers.NewValidActualLRP("some-actual", 0)

		desiredCounter := 0
		fakeEventSource.NextStub = func() (models.Event, error) {
			desiredCounter += 1
			if desiredCounter > 2 {
				return nil, io.EOF
			}
			return models.NewDesiredLRPCreatedEvent(lrp), nil
		}

		actualCounter := 0
		fakeInstanceEventSource.NextStub = func() (models.Event, error) {
			actualCounter += 1
			if actualCounter > 2 {
				return nil, io.EOF
			}
			return models.NewActualLRPInstanceCreatedEvent(actualLRP), nil
		}
	})

	It("prints a JSON object", func() {
		desiredEvent := models.NewDesiredLRPCreatedEvent(lrp)

		desiredLRPEvent := commands.LRPEvent{
			Type: desiredEvent.EventType(),
			Data: desiredEvent,
		}
		data, err := json.Marshal(desiredLRPEvent)
		Expect(err).NotTo(HaveOccurred())
		desiredData := string(data)

		actualEvent := models.NewActualLRPInstanceCreatedEvent(actualLRP)
		actualLRPEvent := commands.LRPEvent{
			Type: actualEvent.EventType(),
			Data: actualEvent,
		}
		data, err = json.Marshal(actualLRPEvent)
		Expect(err).NotTo(HaveOccurred())
		actualData := string(data)

		err = commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
		Expect(err).NotTo(HaveOccurred())

		stdoutData := stdout.Contents()
		lines := strings.Split(string(stdoutData), "\n")
		Expect(lines).To(HaveLen(5))

		Expect(lines).To(ConsistOf(desiredData, desiredData, actualData, actualData, ""))
	})

	It("closes the event stream", func() {
		err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeEventSource.CloseCallCount()).To(Equal(1))
	})

	Context("when failing to subscribe to events", func() {
		BeforeEach(func() {
			fakeBBSClient.SubscribeToEventsByCellIDReturns(nil, errors.New("failed to connect"))
		})

		It("returns an error", func() {
			err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to connect"))
		})
	})

	Context("when failing to receive an event", func() {
		BeforeEach(func() {
			fakeEventSource.NextStub = nil
			fakeEventSource.NextReturns(nil, errors.New("boom"))
		})

		It("returns an error", func() {
			err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("boom"))
		})
	})
})
