package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const gitLabelKey = "org.opencontainers.image.revision"

type Identifier interface {
	Identify(ctx context.Context, c *object.Commit) (string, error)
}

type ImageIdentifier struct {
	identifier Identifier
}

func NewImageIdentifier(id Identifier) *ImageIdentifier {
	return &ImageIdentifier{identifier: id}
}

func (i ImageIdentifier) FindMostRecentImage(ctx context.Context, r *git.Repository) (string, error) {
	commits, err := r.CommitObjects()
	if err != nil {
		return "", fmt.Errorf("failed to get commit objects from repository: %w", err)
	}

	var foundImage string
	err = commits.ForEach(func(c *object.Commit) error {
		if foundImage != "" {
			return nil // Already found, skip remaining commits
		}
		image, err := i.identifier.Identify(ctx, c)
		if err != nil {
			return err
		}
		if image != "" {
			foundImage = image
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return foundImage, nil
}

func main() {
	ctx := context.Background()

	opts := git.CloneOptions{
		URL:           "https://github.com/bigkevmcd/go-demo",
		ReferenceName: plumbing.NewBranchReferenceName("master"),
	}

	dir, err := os.MkdirTemp("", "gitimages")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	log.Printf("cloning repo to %q", dir)

	repo, err := git.PlainClone(dir, false, &opts)
	if err != nil {
		log.Fatal(err)
	}

	labelIdentifier, err := NewLabelIdentifier(ctx, "bigkevmcd/go-demo", gitLabelKey)
	// tagIdentifier, err := NewTagPrefixIdentifier("bigkevmcd/image-updater", "sha-")
	if err != nil {
		log.Fatal(err)
	}

	image, err := NewImageIdentifier(labelIdentifier).FindMostRecentImage(ctx, repo)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("identified %q as the most recent image", image)
}
