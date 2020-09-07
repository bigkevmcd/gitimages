package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Identifier interface {
	Identify(c *object.Commit) (string, error)
}

type ImageIdentifier struct {
	identifier Identifier
}

type TagPrefixIdentifier struct {
	Prefix string
	Tags   []string
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

	return TagPrefixIdentifier{Prefix: prefix, Tags: tags}, nil
}

func (t TagPrefixIdentifier) Identify(c *object.Commit) (string, error) {
	for _, tag := range t.Tags {
		if strings.HasPrefix(c.Hash.String(), strings.TrimPrefix(tag, t.Prefix)) {
			return tag, nil
		}
	}
	return "", nil
}

func NewImageIdentifier(id Identifier) *ImageIdentifier {
	return &ImageIdentifier{identifier: id}
}

func (i ImageIdentifier) FindMostRecentImage(r *git.Repository) (string, error) {
	commits, err := r.CommitObjects()
	if err != nil {
		return "", fmt.Errorf("failed to get commit objects from repository: %w", err)
	}

	foundImage := ""
	foundErr := errors.New("marker error")
	err = commits.ForEach(func(c *object.Commit) error {
		image, err := i.identifier.Identify(c)
		if err != nil {
			return err
		}
		if image != "" {
			foundImage = image
			return foundErr
		}
		return nil
	})
	if err != foundErr {
		return "", err
	}
	return foundImage, nil
}

func main() {
	opts := git.CloneOptions{
		URL:           "https://github.com/gitops-tools/image-updater",
		ReferenceName: plumbing.NewBranchReferenceName("main"),
	}

	dir, err := ioutil.TempDir("", "gitimages")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	log.Printf("cloning repo to %q", dir)

	repo, err := git.PlainClone(dir, false, &opts)
	if err != nil {
		log.Fatal(err)
	}

	tagIdentifier, err := NewTagPrefixIdentifier("bigkevmcd/image-updater", "sha-")
	if err != nil {
		log.Fatal(err)
	}

	image, err := NewImageIdentifier(tagIdentifier).FindMostRecentImage(repo)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("identified %q as the most recent image", image)
}
