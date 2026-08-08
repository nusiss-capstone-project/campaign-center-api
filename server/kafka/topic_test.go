package kafka

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefixedTopic(t *testing.T) {
	t.Setenv("KAFKA_TOPIC_PREFIX", "")
	require.Equal(t, "task.events.completed", PrefixedTopic("task.events.completed"))

	t.Setenv("KAFKA_TOPIC_PREFIX", "dev.")
	require.Equal(t, "dev.task.events.completed", PrefixedTopic("task.events.completed"))
	require.Equal(t, "", PrefixedTopic(""))
	require.Equal(t, []string{"dev.a", "dev.b"}, PrefixedTopics([]string{"a", "b"}))
}

func TestLogicalTopic(t *testing.T) {
	t.Setenv("KAFKA_TOPIC_PREFIX", "dev.")
	require.Equal(t, "task.events.completed", LogicalTopic("dev.task.events.completed"))
	require.Equal(t, "other.topic", LogicalTopic("other.topic"))

	require.NoError(t, os.Unsetenv("KAFKA_TOPIC_PREFIX"))
	require.Equal(t, "dev.task.events.completed", LogicalTopic("dev.task.events.completed"))
}
