package flagsmithapi

import (
	"fmt"
)

type FeatureNotFoundError struct {
	featureUUID string
	featureName string
	projectUUID string
}
type FeatureStateNotFoundError struct {
	featureStateUUID string
}
type SegmentNotFoundError struct {
	segmentUUID string
	segmentName string
	projectUUID string
}
type FeatureMVOptionNotFoundError struct {
	featureMVOptionUUID string
}
type UserNotFoundError struct {
	email string
}

// MultipleFeaturesFoundError is returned when a lookup by name matches more than
// one feature. Feature names are unique within a project, so this is defensive.
type MultipleFeaturesFoundError struct {
	featureName string
	projectUUID string
	count       int
}

// MultipleSegmentsFoundError is returned when a lookup by name matches more than
// one segment. Unlike features, segment names are not unique within a project.
type MultipleSegmentsFoundError struct {
	segmentName string
	projectUUID string
	count       int
}

func (e FeatureNotFoundError) Error() string {
	if e.featureName != "" {
		return fmt.Sprintf("flagsmithapi: feature named '%s' not found in project '%s'", e.featureName, e.projectUUID)
	}
	return fmt.Sprintf("flagsmithapi: feature '%s' not found", e.featureUUID)
}

func (e SegmentNotFoundError) Error() string {
	if e.segmentName != "" {
		return fmt.Sprintf("flagsmithapi: segment named '%s' not found in project '%s'", e.segmentName, e.projectUUID)
	}
	return fmt.Sprintf("flagsmithapi: segment '%s' not found", e.segmentUUID)
}

func (e MultipleFeaturesFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: found %d features named '%s' in project '%s', expected exactly one",
		e.count, e.featureName, e.projectUUID)
}

func (e MultipleSegmentsFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: found %d segments named '%s' in project '%s', expected exactly one",
		e.count, e.segmentName, e.projectUUID)
}

func (e FeatureStateNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: feature state '%s' not found", e.featureStateUUID)
}

func (e FeatureMVOptionNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: feature mv option '%s' not found", e.featureMVOptionUUID)
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: user with email '%s' not found", e.email)
}
