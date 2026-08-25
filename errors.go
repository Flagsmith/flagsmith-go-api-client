package flagsmithapi

import (
	"fmt"
	"strings"
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
type FeatureMVOptionNotFoundError struct {
	featureMVOptionUUID string
}
type UserNotFoundError struct {
	email string
}

// MetadataFieldNotFoundError is returned when no custom field with the given name is
// visible to the project. Fields are exported so that callers can build their own
// messages without parsing this one.
type MetadataFieldNotFoundError struct {
	Name       string
	Entity     string
	KnownNames []string
}

// MetadataFieldNotBoundError is returned when a custom field exists in the project but
// is not enabled for the entity being written to.
type MetadataFieldNotBoundError struct {
	Name   string
	Entity string
}

// MetadataFieldInvalidValueError is returned when a value does not match its custom
// field's configured type.
type MetadataFieldInvalidValueError struct {
	Name      string
	FieldType string
	Value     string
	Reason    string
}

func (e MetadataFieldNotFoundError) Error() string {
	if len(e.KnownNames) == 0 {
		return fmt.Sprintf("flagsmithapi: no custom fields are enabled for %ss in this project", e.Entity)
	}
	return fmt.Sprintf("flagsmithapi: no custom field named '%s' is available for %ss in this project (available: %s)",
		e.Name, e.Entity, strings.Join(e.KnownNames, ", "))
}

func (e MetadataFieldNotBoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: the custom field '%s' exists but is not enabled for %ss",
		e.Name, e.Entity)
}

func (e MetadataFieldInvalidValueError) Error() string {
	return fmt.Sprintf("flagsmithapi: invalid value for custom field '%s': %s", e.Name, e.Reason)
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

func (e FeatureMVOptionNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: feature mv option '%s' not found", e.featureMVOptionUUID)
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("flagsmithapi: user with email '%s' not found", e.email)
}
