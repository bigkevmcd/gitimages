package main

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type TagPrefixIdentifier struct {
	Prefix string
	tags   []string
}

func NewTagPrefixIdentifier(image, prefix string) (TagPrefixIdentifier, error) {
	sourceRepo, err := name.NewRepository(image)
	if err != nil {
		return TagPrefixIdentifier{}, fmt.Errorf("unable to parse image %q: %w", image, err)
	}

	tags, err := remote.List(sourceRepo)
	if err != nil {
		return TagPrefixIdentifier{}, fmt.Errorf("unable to get tags for %q: %w", image, err)
	}

	return TagPrefixIdentifier{Prefix: prefix, tags: tags}, nil
}

func (t TagPrefixIdentifier) Identify(c *object.Commit) (string, error) {
	for _, tag := range t.tags {
		if strings.HasPrefix(c.Hash.String(), strings.TrimPrefix(tag, t.Prefix)) {
			return tag, nil
		}
	}
	return "", nil
}
