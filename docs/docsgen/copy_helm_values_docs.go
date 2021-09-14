package docsgen

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/google/go-github/v32/github"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/stoewer/go-strcase"
)

var (
	helmDocsDir = "content/reference/helm"

	ossFileMapping = map[string]string{
		"codegen/helm/gloo_mesh_helm_values_reference.md":  "%s/%s/gloo_mesh.md",
		"codegen/helm/cert_agent_helm_values_reference.md": "%s/%s/cert_agent.md",
	}

	helmValuesIndex = `
---
title: "%s"
description: Reference for Helm values. 
weight: 2
---

The following pages provide Helm value reference documentation for the open source version of Gloo Mesh.

{{%% children description="true" %%}}
`
	numberMatcher = regexp.MustCompile("[0-9]+")
)

func copyHelmValuesDocsForAllCharts(client *github.Client, rootDir string) error {
	// flush root directory for idempotence
	helmDocsDir := filepath.Join(rootDir, helmDocsDir)
	os.RemoveAll(helmDocsDir)
	os.MkdirAll(helmDocsDir, 0755)

	// create root index
	if err := createFileIfNotExists(helmDocsDir+"/"+"_index.md", fmt.Sprintf(helmValuesIndex, "Helm Values Reference")); err != nil {
		return eris.Errorf("error creating Helm values index file: %v", err)
	}

	// Gloo Mesh OSS
	if err := copyHelmValuesDocsForComponent(
		client,
		rootDir,
		"Gloo Mesh",
		GlooMeshRepoName,
		"v1.0.0",
		ossFileMapping,
	); err != nil {
		return err
	}

	return nil
}

// fetch Helm Values documentation from repo up to and including the version specified by earliestVerison
// fileMapping specifies a mapping from the file path in the origin repo to the file path in this repo
func copyHelmValuesDocsForComponent(
	client *github.Client,
	rootDir string,
	componentName string,
	repoName string,
	earliestVersion string,
	fileMapping map[string]string,
) error {
	// flush directory for idempotence
	helmDocsDir := filepath.Join(rootDir, helmDocsDir, strcase.SnakeCase(componentName))
	os.RemoveAll(helmDocsDir)
	os.MkdirAll(helmDocsDir, 0755)

	if err := createFileIfNotExists(helmDocsDir+"/"+"_index.md", fmt.Sprintf(helmValuesIndex, componentName)); err != nil {
		return eris.Errorf("error creating Helm values index file: %v", err)
	}

	// include Helm values docs for all versions > earliestVersion
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		GithubOrg,
		repoName,
		&github.ListOptions{Page: 0, PerPage: 1000000},
	)
	if err != nil {
		return eris.Errorf("error listing releases: %v", err)
	}

	// the github API returns releases sorted by release date, so we need to sort by version in order to enforce the earliest version lower bound
	var versions []*semver.Version
	for _, release := range releases {
		tagName := release.GetTagName()
		version, err := semver.NewVersion(tagName)
		if err != nil {
			return err
		}
		var modifiedVersion semver.Version
		if version.Prerelease() != "" {
			// semver's comparison function will not put 'beta9' ahead of 'beta10', so we modify the
			// prerelease text to just the number in the prerelease tag.
			match := numberMatcher.FindAllString(version.Prerelease(), -1)
			modifiedVersion, err = version.SetPrerelease(match[0])
			if err != nil {
				return err
			}
			versions = append(versions, &modifiedVersion)
		} else {
			versions = append(versions, version)
		}
	}
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	latestPerMinorVersions := getLatestPerMinorVersion(versions)

	earliestVersionSemver, err := semver.NewVersion(earliestVersion)
	if err != nil {
		return err
	}
	latestVersionSemver, err := semver.NewVersion(earliestVersion)
	if err != nil {
		return err
	}

	tags := make(map[string]string, 0)
	for _, version := range latestPerMinorVersions {
		tags[version.Original()] = fmt.Sprintf("%d.%d", version.Major(), version.Minor())
		if version.GreaterThan(latestVersionSemver) {
			latestVersionSemver = version
		}
		if version.LessThan(earliestVersionSemver) || version.Equal(earliestVersionSemver) {
			break
		}
	}
	tags[latestVersionSemver.Original()] = "latest"

	for tag, tagPath := range tags {
		contextutils.LoggerFrom(context.Background()).Infof("copying Helm values docs from %s/%s for release %s", GithubOrg, repoName, tag)

		if err := os.Mkdir(helmDocsDir+"/"+tagPath, os.ModePerm); err != nil {
			return eris.Errorf("error creating Helm docs directories: %v", err)
		}

		if err := createFileIfNotExists(helmDocsDir+"/"+tagPath+"/"+"_index.md", fmt.Sprintf(helmValuesIndex, tag)); err != nil {
			return eris.Errorf("error creating Helm values index file: %v", err)
		}

		for src, dest := range fileMapping {
			dest = fmt.Sprintf(dest, helmDocsDir, tagPath)
			if err := copyHelmValuesDocs(client, GithubOrg, repoName, tag, src, dest); err != nil {
				return err
			}
		}
	}

	return nil
}

// returns the latest patch version for each minor version
// expects versions to be sorted in reverse order
func getLatestPerMinorVersion(sortedVersions []*semver.Version) []*semver.Version {
	var latestVersions []*semver.Version

	latestVersionForMinor, _ := semver.NewVersion("1.999999999.0")
	for _, version := range sortedVersions {
		if version.Minor() < latestVersionForMinor.Minor() {
			if version.Prerelease() != "" {
				// semver's comparison function will not put 'beta9' ahead of 'beta10', so we revert the modified
				// prerelease text to the original beta-number pattern.
				origVersion, _ := version.SetPrerelease("beta" + version.Prerelease())
				latestVersions = append(latestVersions, &origVersion)
				latestVersionForMinor = &origVersion
			} else {
				latestVersions = append(latestVersions, version)
				latestVersionForMinor = version
			}
		}
	}

	return latestVersions
}

func copyHelmValuesDocs(client *github.Client, org, repo, tag, path, destinationFile string) error {
	contents, _, resp, err := client.Repositories.GetContents(context.Background(), org, repo, path, &github.RepositoryContentGetOptions{
		Ref: tag,
	})

	// return error if expected doc files aren't found
	if err != nil && resp != nil && resp.StatusCode == 404 {
		// 404 is expected for Helm values docs introduced in later versions of Gloo Mesh
		logrus.Warnf("missing Helm values file \"%s\" for version \"%s\"", path, tag)
		return nil
	} else if err != nil {
		return eris.Errorf("error fetching Helm values doc: %v", err)
	}

	decodedContents, err := base64.StdEncoding.DecodeString(*contents.Content)
	if err != nil {
		return eris.Errorf("error fetching Helm values doc: %v", err)
	}

	return createFileIfNotExists(destinationFile, string(decodedContents))
}

// create file with contents, create the file if it doesn't exist
func createFileIfNotExists(fname, contents string) error {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return eris.Errorf("error copying Helm values reference doc: %v", err)
	}

	if _, err = f.Write([]byte(contents)); err != nil {
		return err
	}

	return nil
}
