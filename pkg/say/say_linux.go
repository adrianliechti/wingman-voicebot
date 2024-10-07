//go:build linux

package say

import (
	"errors"
)

func Say(text, language string) error {
	return errors.ErrUnsupported
}
