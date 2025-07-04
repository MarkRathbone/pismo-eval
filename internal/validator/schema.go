package validator

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func Validate(payload []byte) error {
	schemaLoader := gojsonschema.NewReferenceLoader("file://schema/event_schema.json")
	documentLoader := gojsonschema.NewBytesLoader(payload)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errs := ""
		for _, desc := range result.Errors() {
			errs += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("schema validation errors:\n%s", errs)
	}

	return nil
}
