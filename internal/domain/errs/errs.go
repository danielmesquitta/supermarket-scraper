package errs

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
)

type ErrType string

const (
	ErrTypeUnknown                      ErrType = "unknown"
	ErrTypeFailedProcessingProductsPage ErrType = "failed_processing_products_page"
	ErrTypeFailedProcessingCategoryPage ErrType = "failed_processing_category_page"
)

type Err struct {
	Message    string
	StackTrace string
	Type       ErrType
}

// New creates a new Err instance from anything,
// and sets the stack trace.
func New(err any, types ...ErrType) *Err {
	errType := ErrTypeUnknown
	if len(types) > 0 {
		errType = types[0]
	}

	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *Err:
		if v.Type == "" {
			v.Type = errType
		}
		return v

	case error:
		return &Err{
			Message:    v.Error(),
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}

	case string:
		return &Err{
			Message:    v,
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}

	case []byte:
		return &Err{
			Message:    string(v),
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}

	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return &Err{
				Message:    fmt.Sprintf("unsupported err type %T: %+v", v, err),
				StackTrace: string(debug.Stack()),
				Type:       errType,
			}
		}
		return &Err{
			Message:    string(jsonData),
			StackTrace: string(debug.Stack()),
			Type:       errType,
		}
	}
}

func (e *Err) Error() string {
	return e.Message
}

var _ error = (*Err)(nil)
