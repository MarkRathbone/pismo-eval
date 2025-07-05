package validator

import (
	"fmt"
	"log"

	"github.com/xeipuuv/gojsonschema"
)

var schema *gojsonschema.Schema

func init() {
	loader := gojsonschema.NewReferenceLoader("file://schema/event_schema.json")
	var err error
	schema, err = gojsonschema.NewSchema(loader)
	if err != nil {
		log.Fatalf("unable to load schema: %v", err)
	}
}

func Validate(payload []byte) error {
	documentLoader := gojsonschema.NewBytesLoader(payload)

	result, err := schema.Validate(documentLoader)
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
