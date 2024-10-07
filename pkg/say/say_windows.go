//go:build windows

package say

import (
	"errors"
)

func Say(text, language string) error {
	return errors.ErrUnsupported
}
