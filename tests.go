package pgpkg

import (
	"errors"
	"os"
)

type Tests struct {
	Bundle
}

func (p *Package) loadTests(path string) (*Tests, error) {
	bundle, err := p.loadBundle(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Tests{}, nil
		}

		return nil, err
	}

	tests := &Tests{
		Bundle: *bundle,
	}

	return tests, nil
}
