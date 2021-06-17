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

package integration

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/docker/docker/client"

	// latestV1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/ko"
	latestV1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/ko/schema"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestBuildAndSideloadKoImage(t *testing.T) {
	exampleDir, err := koExampleDir()
	if err != nil {
		t.Fatalf("could not get ko example app dir: %+v", err)
	}
	imageNameWithTag := "gcr.io/project-id/skaffold-ko:tag"
	wantImageID := "imageID"

	api := (&testutil.FakeAPIClient{}).Add(imageNameWithTag, wantImageID)
	localDocker := fakeLocalDaemon(api)
	pushImages := false
	b := ko.NewArtifactBuilder(localDocker, pushImages)

	artifact := &latestV1.Artifact{
		ImageName: "ko://github.com/GoogleContainerTools/skaffold/examples/ko",
		Workspace: exampleDir,
		ArtifactType: latestV1.ArtifactType{
			KoArtifact: &latestV1.KoArtifact{},
		},
		Dependencies: []*latestV1.ArtifactDependency{},
	}
	var imageNameBuffer bytes.Buffer
	imageID, err := b.Build(context.TODO(), &imageNameBuffer, artifact, imageNameWithTag)
	if err != nil {
		t.Fatalf("error during build: %+v", err)
	}

	if imageID != wantImageID {
		t.Errorf("got image ID %s, wanted %s", imageID, wantImageID)
	}
	imageName := imageNameBuffer.String()
	wantImageNamePrefix := "gcr.io/project-id/skaffold-ko:"
	if !strings.HasPrefix(imageName, wantImageNamePrefix) {
		t.Errorf("got image name %s, wanted image name with prefix %s", imageName, wantImageNamePrefix)
	}
}

func koExampleDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not get current filename")
	}
	basepath := filepath.Dir(filename)
	exampleDir, err := filepath.Abs(filepath.Join(basepath, "examples", "ko"))
	if err != nil {
		return "", fmt.Errorf("could not get absolute path of example from basepath %q: %w", basepath, err)
	}
	return exampleDir, nil
}

func fakeLocalDaemon(api client.CommonAPIClient) docker.LocalDaemon {
	return docker.NewLocalDaemon(api, nil, false, nil)
}
