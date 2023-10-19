package webview2

import (
	"github.com/zellyn/kooky/internal/chrome/find"
)

func FindWebView2CookieStoreFiles(rootsFunc func() ([]string, error)) ([]*find.ChromeCookieStoreFile, error) {
	return find.FindWebView2CookieStoreFiles(rootsFunc)
}
