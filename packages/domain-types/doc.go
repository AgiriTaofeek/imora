//go:generate sh -c "go-jsonschema --package eventschemas --struct-name-from-title --capitalization ID ../event-schemas/schemas/*.schema.json > gen_event_schemas.go"
package eventschemas
