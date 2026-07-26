package infrastructurenodeshared

import (
	"fmt"
	"strings"
)

func GenericPubTopic(suffix string) string {
	return fmt.Sprintf("/pub/%s", strings.TrimPrefix(suffix, "/"))
}

func GenericSubTopic(suffix string) string {
	return fmt.Sprintf("/sub/%s", strings.TrimPrefix(suffix, "/"))
}

func NodePubTopic(nodeDeviceId string, suffix string) string {
	return fmt.Sprintf("/pub/%s/%s", nodeDeviceId, strings.TrimPrefix(suffix, "/"))
}

func NodeSubTopic(nodeDeviceId string, suffix string) string {
	return fmt.Sprintf("/sub/%s/%s", nodeDeviceId, strings.TrimPrefix(suffix, "/"))
}
