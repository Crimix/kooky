package webview2

import (
	"github.com/zellyn/kooky/internal/chrome"
	"github.com/zellyn/kooky/internal/chrome/find"
	"github.com/zellyn/kooky/internal/cookies"
)

func FindWebView2CookieStoreFiles(rootsFunc func() ([]string, error)) ([]*find.ChromeCookieStoreFile, error) {
	return find.FindWebView2CookieStoreFiles(rootsFunc)
}

func CreateCookieStore(file *find.ChromeCookieStoreFile) *cookies.CookieJar {
	return &cookies.CookieJar{
		CookieStore: &chrome.CookieStore{
			DefaultCookieStore: cookies.DefaultCookieStore{
				BrowserStr:           file.Browser,
				ProfileStr:           file.Profile,
				OSStr:                file.OS,
				IsDefaultProfileBool: file.IsDefaultProfile,
				FileNameStr:          file.Path,
			},
		},
	}
}
