package appmock

import (
	"testing"

	"github.com/IBM/sarama"
)

// initMockBroker initializes a mock Kafka broker with predefined responses
func InitMockBroker(t *testing.T, topic string) *sarama.MockBroker {
	// Initialize a mock Kafka broker with predefined responses
	topics := []string{topic}
	mockBroker := sarama.NewMockBroker(t, 0)

	mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
			SetLeader(topics[0], 0, mockBroker.BrokerID()).
			SetController(mockBroker.BrokerID()),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset(topics[0], 0, sarama.OffsetOldest, 0).
			SetOffset(topics[0], 0, sarama.OffsetNewest, 1),
		"FindCoordinatorRequest": sarama.NewMockFindCoordinatorResponse(t).
			SetCoordinator(sarama.CoordinatorGroup, "tests", mockBroker),
		"HeartbeatRequest": sarama.NewMockHeartbeatResponse(t),
		"JoinGroupRequest": sarama.NewMockSequence(
			sarama.NewMockJoinGroupResponse(t).SetError(sarama.ErrOffsetsLoadInProgress),
			sarama.NewMockJoinGroupResponse(t).SetGroupProtocol(sarama.RangeBalanceStrategyName),
		),
		"SyncGroupRequest": sarama.NewMockSequence(
			sarama.NewMockSyncGroupResponse(t).SetError(sarama.ErrOffsetsLoadInProgress),
			sarama.NewMockSyncGroupResponse(t).SetMemberAssignment(
				&sarama.ConsumerGroupMemberAssignment{
					Version: 0,
					Topics: map[string][]int32{
						topics[0]: {0},
					},
				}),
		),
		"OffsetFetchRequest": sarama.NewMockOffsetFetchResponse(t).SetOffset(
			"tests", topics[0], 0, 0, "", sarama.ErrNoError,
		).SetError(sarama.ErrNoError),
		"FetchRequest": sarama.NewMockSequence(
			sarama.NewMockFetchResponse(t, 1).
				SetMessage(topics[0], 0, 0, sarama.StringEncoder("foo")),
		),
	})
	return mockBroker
}
