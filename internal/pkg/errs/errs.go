package errs

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
)

type Err struct {
	Message    string
	StackTrace string
}

// New creates a new Err instance from anything,
// and sets the stack trace.
func New(err any) *Err {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case *Err:
		return v
	case error:
		return &Err{
			Message:    v.Error(),
			StackTrace: string(debug.Stack()),
		}
	case string:
		return &Err{
			Message:    v,
			StackTrace: string(debug.Stack()),
		}
	case []byte:
		return &Err{
			Message:    string(v),
			StackTrace: string(debug.Stack()),
		}

	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return &Err{
				Message:    fmt.Sprintf("unsupported err type %T: %+v", v, err),
				StackTrace: string(debug.Stack()),
			}
		}
		return &Err{
			Message:    string(jsonData),
			StackTrace: string(debug.Stack()),
		}
	}
}

func (e *Err) Error() string {
	return e.Message
}

var _ error = (*Err)(nil)
