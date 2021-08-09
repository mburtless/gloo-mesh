package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func GetLatestChartVersion(repoURI, chartName string, stable bool) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version semver.Version) bool {
		// Do not allow prereleaes if stable is true.
		return !stable || version.Prerelease() == ""
	})
}

func GetLatestChartMinorVersion(repoURI, chartName string, stable bool, major, minor int64) (string, error) {
	return getLatestChartVersion(repoURI, chartName, func(version semver.Version) bool {
		// Compatible versions will have the given major and minor version.
		// Do not allow prereleaes if stable is true.
		return version.Major() == major && version.Minor() == minor &&
			(!stable || version.Prerelease() == "")
	})
}

func getLatestChartVersion(
	repoURI, chartName string,
	isVersionCompatible func(version semver.Version) bool,
) (string, error) {
	versions, err := getChartVersions(repoURI, chartName)
	if err != nil {
		return "", nil
	}
	versions = sortPrereleaseVersions(versions) // Sort from newest to oldest
	logrus.Debugf("available versions: %v", versions)

	for _, version := range versions {
		if isVersionCompatible(*version) {
			logrus.Debugf("installing chart version %s", version.Original())
			return version.Original(), nil
		}
	}

	return "", eris.New("compatible chart version not found")
}

func getChartVersions(repoURI, chartName string) (semver.Collection, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s/index.yaml", repoURI, chartName))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, eris.Wrap(err, "unable to read response body")
	}
	if res.StatusCode != http.StatusOK {
		logrus.Debug(string(b))
		return nil, eris.Errorf("invalid response from the Helm repository: %d %s", res.StatusCode, res.Status)
	}
	index := struct {
		Entries map[string][]struct {
			Version string `yaml:"version"`
		} `yaml:"entries"`
	}{}
	if err := yaml.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	chartReleases, ok := index.Entries[chartName]
	if !ok {
		logrus.Debug(string(b))
		return nil, eris.Errorf("chart not found in index: %s", chartName)
	}
	versions := make(semver.Collection, 0, len(chartReleases))
	for _, release := range chartReleases {
		version, err := semver.NewVersion(release.Version)
		if err != nil {
			logrus.Warnf("invalid release version: %s", release.Version)
			continue
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// semver's comparison function will not put 'beta9' ahead of 'beta10', so we modify the
// prerelease text to the semver-accepted pattern with a '.' in front of the number.
func sortPrereleaseVersions(versions semver.Collection) semver.Collection {
	var modifiedVersions semver.Collection
	var sortedVersions semver.Collection

	for _, v := range versions {
		modifiedV, err := v.SetPrerelease(strings.ReplaceAll(v.Prerelease(), "beta", "beta."))
		if err != nil {
			logrus.Warnf("invalid release version: %s", v)
			continue
		}
		modifiedVersions = append(modifiedVersions, &modifiedV)
	}

	sort.Sort(sort.Reverse(modifiedVersions)) // Sort from newest to oldest

	for _, v := range modifiedVersions {
		originalV, err := v.SetPrerelease(strings.ReplaceAll(v.Prerelease(), "beta.", "beta"))
		if err != nil {
			logrus.Warnf("invalid release version: %s", v)
			continue
		}
		sortedVersions = append(sortedVersions, &originalV)
	}
	return sortedVersions
}
