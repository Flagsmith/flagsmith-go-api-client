package flagsmithapi

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// MetadataFieldValueMaxLength mirrors FIELD_VALUE_MAX_LENGTH in
// api/metadata/models.py.
const MetadataFieldValueMaxLength = 2000

// metadataFieldsPageLimit guards against looping forever on a bad `next` link.
const metadataFieldsPageLimit = 100

// Metadata is a single custom field value attached to an entity.
//
// FieldValue is always a string on the wire, whatever the custom field's configured
// type. ModelField is a MetadataModelField ID, not a MetadataField ID.
type Metadata struct {
	ID         int64  `json:"id,omitempty"`
	ModelField int64  `json:"model_field"`
	FieldValue string `json:"field_value"`
}

// MetadataModelField binds a custom field to a Django content type, i.e. to features,
// segments or environments. Its ID is what Metadata.ModelField refers to.
type MetadataModelField struct {
	ID          int64 `json:"id"`
	ContentType int64 `json:"content_type"`
}

// MetadataField is a custom field definition. Fields belong to an organisation and may
// optionally be scoped to a single project.
type MetadataField struct {
	ID           int64                `json:"id"`
	Name         string               `json:"name"`
	Type         string               `json:"type"`
	Description  *string              `json:"description"`
	Organisation int64                `json:"organisation"`
	Project      *int64               `json:"project"`
	ModelFields  []MetadataModelField `json:"model_fields"`
}

// ValidateValue mirrors MetadataField.is_field_value_valid in
// api/metadata/models.py. The API remains authoritative: unknown field types are
// accepted here and left for the server to reject.
func (f MetadataField) ValidateValue(value string) error {
	if value == "" {
		// field_value is a TextField with blank=False, so DRF rejects empty values.
		return fmt.Errorf("value for custom field '%s' must not be empty", f.Name)
	}
	if len(value) > MetadataFieldValueMaxLength {
		return fmt.Errorf("value for custom field '%s' is %d characters, the maximum is %d",
			f.Name, len(value), MetadataFieldValueMaxLength)
	}

	switch f.Type {
	case "int":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("custom field '%s' is of type 'int', but got '%s'", f.Name, value)
		}
	case "bool":
		lowered := strings.ToLower(value)
		if lowered != "true" && lowered != "false" {
			return fmt.Errorf("custom field '%s' is of type 'bool', but got '%s'", f.Name, value)
		}
	case "url":
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("custom field '%s' is of type 'url', but got '%s'", f.Name, value)
		}
	}

	return nil
}

// ContentTypeInfo identifies a Django content type. IDs are specific to a Flagsmith
// installation and must never be hardcoded.
type ContentTypeInfo struct {
	ID       int64  `json:"id"`
	AppLabel string `json:"app_label"`
	Model    string `json:"model"`
}

// MetadataEntity is an entity that custom fields can be attached to.
type MetadataEntity string

const (
	MetadataEntityFeature     MetadataEntity = "feature"
	MetadataEntitySegment     MetadataEntity = "segment"
	MetadataEntityEnvironment MetadataEntity = "environment"
)

// metadataEntityAppLabels are the Django app labels each entity's content type is
// expected to live in, used to disambiguate content types with the same model name.
var metadataEntityAppLabels = map[MetadataEntity]string{
	MetadataEntityFeature:     "features",
	MetadataEntitySegment:     "segments",
	MetadataEntityEnvironment: "environments",
}

// GetProjectMetadataFields returns every custom field definition visible to a project:
// the project's own fields plus the inherited organisation level fields.
//
// NOTE: the endpoint's `entity` query param is deliberately not used. It filters which
// *fields* are returned but does NOT filter each field's nested `model_fields`, so
// callers have to match on content type regardless. Fetching everything lets a single
// response serve all three entity types and produce precise "this field exists but is
// not enabled for features" errors.
func (c *Client) GetProjectMetadataFields(projectID int64) ([]MetadataField, error) {
	requestURL := fmt.Sprintf("%s/projects/%d/metadata/fields/", c.baseURL, projectID)
	fields := []MetadataField{}

	for page := 1; page <= metadataFieldsPageLimit; page++ {
		result := struct {
			Next    *string         `json:"next"`
			Results []MetadataField `json:"results"`
		}{}

		resp, err := c.client.R().
			SetQueryParams(map[string]string{
				// Without this, organisation level fields -- where custom fields
				// created in the UI usually live -- are omitted entirely.
				"include_organisation": "true",
				"page":                 strconv.Itoa(page),
			}).
			SetResult(&result).
			Get(requestURL)

		if err != nil {
			return nil, err
		}

		if !resp.IsSuccess() {
			return nil, fmt.Errorf("flagsmithapi: Error getting metadata fields: %s", resp)
		}

		fields = append(fields, result.Results...)

		if result.Next == nil || *result.Next == "" {
			return fields, nil
		}
	}

	return nil, fmt.Errorf("flagsmithapi: Error getting metadata fields: more than %d pages of results",
		metadataFieldsPageLimit)
}

// GetSupportedMetadataContentTypes returns the content types that custom fields can be
// bound to for the given organisation.
func (c *Client) GetSupportedMetadataContentTypes(organisationID int64) ([]ContentTypeInfo, error) {
	requestURL := fmt.Sprintf("%s/organisations/%d/metadata-model-fields/supported-content-types/",
		c.baseURL, organisationID)

	contentTypes := []ContentTypeInfo{}
	resp, err := c.client.R().SetResult(&contentTypes).Get(requestURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting supported metadata content types: %s", resp)
	}

	return contentTypes, nil
}

// contentTypeIDForEntity resolves the content type ID for an entity, preferring an exact
// (app_label, model) match and falling back to a unique match on the model name alone.
func contentTypeIDForEntity(contentTypes []ContentTypeInfo, entity MetadataEntity) (int64, error) {
	var candidates []ContentTypeInfo

	for _, contentType := range contentTypes {
		if contentType.Model != string(entity) {
			continue
		}
		if contentType.AppLabel == metadataEntityAppLabels[entity] {
			return contentType.ID, nil
		}
		candidates = append(candidates, contentType)
	}

	switch len(candidates) {
	case 0:
		return 0, fmt.Errorf("flagsmithapi: no content type found for '%s'; custom fields may not be supported by this Flagsmith version", entity)
	case 1:
		return candidates[0].ID, nil
	default:
		return 0, fmt.Errorf("flagsmithapi: found %d content types named '%s'; cannot determine which one custom fields apply to",
			len(candidates), entity)
	}
}

// MetadataFieldResolver maps custom field names to MetadataModelField IDs for a single
// project and entity type, and back again.
type MetadataFieldResolver struct {
	entity        MetadataEntity
	contentTypeID int64

	// nameToModelFieldID only covers fields bound to this entity.
	nameToModelFieldID map[string]int64
	nameToField        map[string]MetadataField
	// modelFieldIDToName covers every model field in the project, whatever entity it
	// is bound to, so that reverse lookups never fail spuriously.
	modelFieldIDToName map[int64]string
	// unboundNames are fields that exist in the project but are not enabled for this
	// entity.
	unboundNames map[string]struct{}
}

func newMetadataFieldResolver(entity MetadataEntity, contentTypeID int64, fields []MetadataField) *MetadataFieldResolver {
	resolver := &MetadataFieldResolver{
		entity:             entity,
		contentTypeID:      contentTypeID,
		nameToModelFieldID: map[string]int64{},
		nameToField:        map[string]MetadataField{},
		modelFieldIDToName: map[int64]string{},
		unboundNames:       map[string]struct{}{},
	}

	for _, field := range fields {
		bound := false
		for _, modelField := range field.ModelFields {
			resolver.modelFieldIDToName[modelField.ID] = field.Name
			if modelField.ContentType == contentTypeID {
				resolver.nameToModelFieldID[field.Name] = modelField.ID
				resolver.nameToField[field.Name] = field
				bound = true
			}
		}
		if !bound {
			resolver.unboundNames[field.Name] = struct{}{}
		}
	}

	return resolver
}

// ModelFieldID resolves a custom field name to the model field ID to write against.
//
// It returns a MetadataFieldNotBoundError if the field exists in the project but is not
// enabled for this entity, and a MetadataFieldNotFoundError if it does not exist at all.
func (r *MetadataFieldResolver) ModelFieldID(name string) (int64, error) {
	if modelFieldID, ok := r.nameToModelFieldID[name]; ok {
		return modelFieldID, nil
	}
	if _, ok := r.unboundNames[name]; ok {
		return 0, MetadataFieldNotBoundError{Name: name, Entity: string(r.entity)}
	}
	return 0, MetadataFieldNotFoundError{Name: name, Entity: string(r.entity), KnownNames: r.BoundNames()}
}

// Name reverse resolves a model field ID. ok is false for IDs the resolver has never
// seen, for example when the custom field has been deleted since the value was written.
func (r *MetadataFieldResolver) Name(modelFieldID int64) (string, bool) {
	name, ok := r.modelFieldIDToName[modelFieldID]
	return name, ok
}

// BoundNames returns the sorted names of the custom fields available for this entity.
func (r *MetadataFieldResolver) BoundNames() []string {
	names := make([]string, 0, len(r.nameToModelFieldID))
	for name := range r.nameToModelFieldID {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Field returns the definition of a custom field bound to this entity.
func (r *MetadataFieldResolver) Field(name string) (MetadataField, bool) {
	field, ok := r.nameToField[name]
	return field, ok
}

// GetMetadataFieldResolver returns a resolver for the given project and entity.
//
// This costs three requests: the project, its organisation's content types, and the
// project's custom field definitions.
func (c *Client) GetMetadataFieldResolver(projectID int64, entity MetadataEntity) (*MetadataFieldResolver, error) {
	project, err := c.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	contentTypes, err := c.GetSupportedMetadataContentTypes(project.Organisation)
	if err != nil {
		return nil, err
	}

	contentTypeID, err := contentTypeIDForEntity(contentTypes, entity)
	if err != nil {
		return nil, err
	}

	fields, err := c.GetProjectMetadataFields(projectID)
	if err != nil {
		return nil, err
	}

	return newMetadataFieldResolver(entity, contentTypeID, fields), nil
}

// BuildMetadata converts a map of custom field names to values into the wire format,
// validating each value against its field's configured type.
//
// A nil or empty map yields a non-nil empty slice, which serialises as `"metadata": []`,
// i.e. an explicit "this entity has no custom field values".
func (c *Client) BuildMetadata(projectID int64, entity MetadataEntity, values map[string]string) (*[]Metadata, error) {
	metadata := []Metadata{}
	if len(values) == 0 {
		return &metadata, nil
	}

	resolver, err := c.GetMetadataFieldResolver(projectID, entity)
	if err != nil {
		return nil, err
	}

	// Sorted so that request bodies are deterministic.
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		modelFieldID, err := resolver.ModelFieldID(name)
		if err != nil {
			return nil, err
		}
		if field, ok := resolver.Field(name); ok {
			if err := field.ValidateValue(values[name]); err != nil {
				return nil, MetadataFieldInvalidValueError{
					Name:      name,
					FieldType: field.Type,
					Value:     values[name],
					Reason:    err.Error(),
				}
			}
		}
		metadata = append(metadata, Metadata{ModelField: modelFieldID, FieldValue: values[name]})
	}

	return &metadata, nil
}

// ResolveMetadataNames converts wire format metadata back into a map of custom field
// names to values.
//
// Model field IDs the resolver does not recognise are skipped and returned in
// unresolved, so that a deleted custom field cannot make a read fail.
func (c *Client) ResolveMetadataNames(projectID int64, entity MetadataEntity, metadata []Metadata) (map[string]string, []int64, error) {
	values := map[string]string{}
	if len(metadata) == 0 {
		return values, nil, nil
	}

	resolver, err := c.GetMetadataFieldResolver(projectID, entity)
	if err != nil {
		return nil, nil, err
	}

	var unresolved []int64
	for _, item := range metadata {
		name, ok := resolver.Name(item.ModelField)
		if !ok {
			unresolved = append(unresolved, item.ModelField)
			continue
		}
		values[name] = item.FieldValue
	}

	return values, unresolved, nil
}
