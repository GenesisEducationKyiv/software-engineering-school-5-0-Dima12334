package clients

import (
	"fmt"
	"io"
)

func closeBody(body io.Closer, errPtr *error) {
	if closeErr := body.Close(); closeErr != nil {
		if *errPtr != nil {
			*errPtr = fmt.Errorf("%w; failed to close response body: %w", *errPtr, closeErr)
		} else {
			*errPtr = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}
}
