/*
Copyright 2021 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ko

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/publish"

	// latestV1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	latestV1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/ko/schema"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/version"
)

// Build builds an artifact with ko
func (b *Builder) Build(ctx context.Context, out io.Writer, a *latestV1.Artifact, fullImageNameWithTag string) (string, error) {
	koBuilder, err := b.newKoBuilder(ctx, a)
	if err != nil {
		return "", fmt.Errorf("error creating ko builder: %w", err)
	}

	koPublisher, err := b.newKoPublisher(fullImageNameWithTag)
	if err != nil {
		return "", fmt.Errorf("error creating ko publisher: %w", err)
	}
	defer koPublisher.Close()

	imageRef, err := b.buildAndPublish(ctx, a.ImageName, koBuilder, koPublisher)
	if err != nil {
		return "", fmt.Errorf("could not build and publish ko image %q: %w", a.ImageName, err)
	}
	fmt.Fprintln(out, imageRef.Name())

	return b.getImageIdentifier(ctx, imageRef, fullImageNameWithTag)
}

func (b *Builder) newKoBuilder(ctx context.Context, a *latestV1.Artifact) (build.Interface, error) {
	bo := &options.BuildOptions{
		BaseImage:        a.KoArtifact.BaseImage,
		ConcurrentBuilds: 1, // TODO(halvards) link to Skaffold concurrent builds?
		Platform:         strings.Join(a.KoArtifact.Platforms, ","),
		UserAgent:        version.UserAgent(),
		WorkingDirectory: a.Workspace,
	}
	return commands.NewBuilder(ctx, bo)
}

func (b *Builder) newKoPublisher(fullImageNameWithTag string) (publish.Interface, error) {
	ref, err := name.ParseReference(fullImageNameWithTag)
	if err != nil {
		return nil, err
	}
	imageNameWithoutTag := ref.Context().Name()
	po := &options.PublishOptions{
		Bare:        true,
		DockerRepo:  imageNameWithoutTag,
		Local:       !b.pushImages,
		LocalDomain: imageNameWithoutTag,
		Push:        b.pushImages,
		Tags:        []string{ref.Identifier()},
		UserAgent:   version.UserAgent(),
	}
	return commands.NewPublisher(po)
}

func getImportPath(imageName string, koBuilder build.Interface) (string, error) {
	if strings.HasPrefix(imageName, `ko://`) {
		return imageName, nil
	}
	return koBuilder.QualifyImport(".")
}

func (b *Builder) buildAndPublish(ctx context.Context, imageName string, koBuilder build.Interface, koPublisher publish.Interface) (name.Reference, error) {
	importpath, err := getImportPath(imageName, koBuilder)
	if err != nil {
		return nil, fmt.Errorf("could not determine Go import path for ko image %q: %w", imageName, err)
	}

	imageMap, err := b.publishImages(ctx, []string{importpath}, koPublisher, koBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to publish ko image: %w", err)
	}
	imageRef, exists := imageMap[importpath]
	if !exists {
		return nil, fmt.Errorf("no built image found for Go import path %q build images: %+v", importpath, imageMap)
	}
	return imageRef, nil
}

func (b *Builder) getImageIdentifier(ctx context.Context, imageRef name.Reference, fullImageNameWithTag string) (string, error) {
	if b.pushImages {
		return imageRef.Identifier(), nil
	}
	imageIdentifier, err := b.localDocker.ImageID(ctx, fullImageNameWithTag)
	if err != nil {
		return "", fmt.Errorf("could not get imageID from local Docker Daemon for image %s: %+v", fullImageNameWithTag, err)
	}
	return imageIdentifier, nil
}
