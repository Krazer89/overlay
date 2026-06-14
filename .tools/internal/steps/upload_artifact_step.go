// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package steps

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-github/v88/github"
	"github.com/krazer89/overlay/.tools/internal/steps/stepshelpers"
)

// UploadArtifactStep takes the provided path and uploads it from the
// container to a GitHub Release asset. This is used to store artifacts that
// are required by ebuilds (e.g., Go dependency archives) and retrieve
// them later.
//
// GitHub configuration is provided by the environment, and authentication
// expects a GITHUB_TOKEN environment variable.
type UploadArtifactStep struct {
	// path is the path to the artifact.
	path string
}

// NewUploadArtifactStep creates a new UploadArtifactStep from the provided input.
func NewUploadArtifactStep(input any) (StepRunner, error) {
	path, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", input)
	}

	return &UploadArtifactStep{path}, nil
}

// Run runs the provided command inside of the step runner.
func (e UploadArtifactStep) Run(ctx context.Context, env Environment) (*StepOutput, error) {
	if !filepath.IsAbs(e.path) {
		e.path = filepath.Join(env.workDir, e.path)
	}

	// Dynamic fallback configuration parsing to prevent structural compilation blockers
	var owner, repo string

	//TODO: Config for non-github
	owner = os.Getenv("GITHUB_OWNER")
	repo = os.Getenv("GITHUB_REPO")

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("github owner or repo was not set via environment (GITHUB_OWNER/GITHUB_REPO)")
	}

	// Authenticate using the standard GITHUB_TOKEN environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	// Instantiating modern go-github v88 client using functional options
	client, err := github.NewClient(github.WithAuthToken(token))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize github client: %w", err)
	}

	// Format the tag name as "package-version" (e.g., "example-2.6.5")
	tagName := fmt.Sprintf("%s-%s", env.in.OriginalEbuild.Name, env.in.LatestVersion)
	uploadFileName := filepath.Base(e.path)

	// 1. Check if the release already exists for this tag
	var release *github.RepositoryRelease
	existingRelease, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tagName)
	if err == nil {
		release = existingRelease
	} else {
		// 2. If it doesn't exist, create a new release
		env.log.With("tag", tagName).Info("release not found, creating a new one")
		newRelease := &github.RepositoryRelease{
			TagName: github.String(tagName),
			Name:    github.String(tagName),
		}
		release, _, err = client.Repositories.CreateRelease(ctx, owner, repo, newRelease)
		if err != nil {
			return nil, fmt.Errorf("failed to create github release: %w", err)
		}
	}

	// 3. Stream the file out of the container
	out, size, wait, err := stepshelpers.StreamFileFromContainer(ctx, env.containerID, e.path)
	if err != nil {
		return nil, fmt.Errorf("failed to stream file from container: %w", err)
	}

	// go-github requires an explicit *os.File. We copy the stream to a temporary host file.
	tmpFile, err := os.CreateTemp("", "overlay-artifact-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for asset upload: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, out); err != nil {
		return nil, fmt.Errorf("failed to buffer container stream to local temp file: %w", err)
	}

	if err := wait(); err != nil {
		return nil, fmt.Errorf("failed to wait for command to finish: %w", err)
	}

	// Seek back to the beginning of the file before uploading
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind local temp file: %w", err)
	}

	env.log.With("file", uploadFileName, "size", size, "release", tagName).Info("uploading artifact to github release")

	// 4. Upload the asset to the release
	opt := &github.UploadOptions{
		Name: uploadFileName,
	}
	// Pass the concrete *os.File directly now
	_, _, err = client.Repositories.UploadReleaseAsset(ctx, owner, repo, release.GetID(), opt, tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload release asset: %w", err)
	}

	env.log.Info("successfully uploaded artifact to github release")

	return nil, nil
}