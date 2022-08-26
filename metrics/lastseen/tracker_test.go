package lastseen

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TrackerTestSuite struct {
	suite.Suite
}

type testExporter struct {
	mock.Mock
}

func (e *testExporter) Export(ctx context.Context, field string, val LastSeen) error {
	args := e.Called(ctx, field, val)
	return args.Error(0)
}

func TestLastseen(t *testing.T) {
	suite.Run(t, new(TrackerTestSuite))
}

func (s TrackerTestSuite) TestExportIsBurstControlled() {
	// Test if a burst of registrations are only exported once
	ctx := context.Background()
	exporterMock := new(testExporter)
	exporterMock.On("Export", ctx, "TestField", mock.AnythingOfType("LastSeen")).Return(nil)

	tracker, err := NewTrackerSingleton(ctx,
		WithFlushIntervalSeconds(1),
		WithTicker(false),
		WithExporter(func() (Exporter, error) { return exporterMock, nil }))

	if err != nil {
		s.Error(err)
	}

	tracker.SetSeen(ctx, "TestField")
	for i := 0; i < 10; i++ {
		tracker.SetSeen(ctx, "TestField")
	}
	// sleep enough to have a flush happen
	time.Sleep(2 * time.Second)
	tracker.SetSeen(ctx, "TestField")

	tracker.WaitForExporters()

	exporterMock.AssertNumberOfCalls(s.T(), "Export", 2)
}
