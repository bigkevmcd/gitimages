package main

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type LabelIdentifier struct {
	Label     string
	tags      []string
	repo      name.Repository
	tagLabels map[string]map[string]string
}

func NewLabelIdentifier(image, label string) (*LabelIdentifier, error) {
	repo, err := name.NewRepository(image)
	if err != nil {
		return &LabelIdentifier{}, fmt.Errorf("unable to parse image %q: %w", image, err)
	}

	tags, err := remote.List(repo)
	if err != nil {
		return &LabelIdentifier{}, fmt.Errorf("unable to get tags for %q: %w", image, err)
	}
	return &LabelIdentifier{
		Label:     label,
		repo:      repo,
		tags:      tags,
		tagLabels: make(map[string]map[string]string),
	}, nil
}

func (t LabelIdentifier) Identify(c *object.Commit) (string, error) {
	for _, tag := range t.tags {
		if t.tagLabels[tag] == nil {
			i, err := remote.Image(t.repo.Tag(tag))
			if err != nil {
				return "", fmt.Errorf("failed to get image details for image %q: %w", t.repo.Tag(tag), err)
			}
			cfg, err := i.ConfigFile()
			if err != nil {
				return "", fmt.Errorf("failed to get config file for image %q: %w", t.repo.Tag(tag), err)
			}
			t.tagLabels[tag] = cfg.Config.Labels
		}
		if l := t.tagLabels[tag]; l != nil {
			if l[t.Label] == c.Hash.String() {
				return tag, nil
			}
		}
	}
	return "", nil
}
