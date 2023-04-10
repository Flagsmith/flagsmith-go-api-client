package flagsmithapi

import (
	"fmt"
)

type FeatureNotFoundError struct {
	featureUUID string
}
type FeatureStateNotFoundError struct {
	featureStateUUID string
}
type SegmentNotFoundError struct {
	segmentUUID string
}

func (e FeatureNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: feature '%s' not found", e.featureUUID)
}

func (e SegmentNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: segment '%s' not found", e.segmentUUID)
}

func (e FeatureStateNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: feature state '%s' not found", e.featureStateUUID)
}
