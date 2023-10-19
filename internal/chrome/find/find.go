package find

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type ChromeCookieStoreFile struct {
	Path             string
	Browser          string
	Profile          string
	OS               string
	IsDefaultProfile bool
}

// chromeRoots and chromiumRoots could be put into the github.com/kooky/browser/{chrome,chromium} packages.
// It might be better though to keep those 2 together here as they are based on the same source.
func FindChromeCookieStoreFiles() ([]*ChromeCookieStoreFile, error) {
	return FindCookieStoreFiles(chromeRoots, `chrome`)
}
func FindChromiumCookieStoreFiles() ([]*ChromeCookieStoreFile, error) {
	return FindCookieStoreFiles(chromiumRoots, `chromium`)
}

func FindBraveCookieStoreFiles() ([]*ChromeCookieStoreFile, error) {
	return FindCookieStoreFiles(braveRoots, `brave`)
}

func FindWebView2CookieStoreFiles(rootsFunc func() ([]string, error)) ([]*ChromeCookieStoreFile, error) {
	return FindCookieStoreFiles(rootsFunc, `webview2`)
}

func FindCookieStoreFiles(rootsFunc func() ([]string, error), browserName string) ([]*ChromeCookieStoreFile, error) {
	if rootsFunc == nil {
		return nil, errors.New(`passed roots function is nil`)
	}
	var files []*ChromeCookieStoreFile
	roots, err := rootsFunc()
	if err != nil {
		return nil, err
	}
	for _, root := range roots {
		localStateBytes, err := ioutil.ReadFile(filepath.Join(root, `Local State`))
		if err != nil {
			continue
		}
		var localState struct {
			Profile struct {
				InfoCache map[string]struct {
					IsUsingDefaultName bool `json:"is_using_default_name"`
					Name               string
				} `json:"info_cache"`
			}
		}
		if err := json.Unmarshal(localStateBytes, &localState); err != nil {
			// fallback - json file exists, json structure unknown
			files = append(
				files,
				[]*ChromeCookieStoreFile{
					{
						Browser:          browserName,
						Profile:          `Profile 1`,
						IsDefaultProfile: true,
						Path:             filepath.Join(root, `Default`, `Network`, `Cookies`), // Chrome 96
						OS:               runtime.GOOS,
					},
					{
						Browser:          browserName,
						Profile:          `Profile 1`,
						IsDefaultProfile: true,
						Path:             filepath.Join(root, `Default`, `Cookies`),
						OS:               runtime.GOOS,
					},
				}...,
			)
			continue

		}
		for profDir, profStr := range localState.Profile.InfoCache {
			files = append(
				files,
				[]*ChromeCookieStoreFile{
					{
						Browser:          browserName,
						Profile:          profStr.Name,
						IsDefaultProfile: profStr.IsUsingDefaultName,
						Path:             filepath.Join(root, profDir, `Network`, `Cookies`), // Chrome 96
						OS:               runtime.GOOS,
					}, {
						Browser:          browserName,
						Profile:          profStr.Name,
						IsDefaultProfile: profStr.IsUsingDefaultName,
						Path:             filepath.Join(root, profDir, `Cookies`),
						OS:               runtime.GOOS,
					},
				}...,
			)
		}
	}
	return files, nil
}
