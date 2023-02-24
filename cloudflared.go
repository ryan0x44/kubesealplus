package main

import (
	"fmt"
	"net/url"

	"github.com/cloudflare/cloudflared/token"
)

func getCloudflareAccessAppInfo(u *url.URL) (isAccessApp bool, appInfo *token.AppInfo, err error) {
	notAccessErr := fmt.Sprintf("failed to find Access application at %s", u.String())
	appInfo, err = token.GetAppInfo(u)
	isAccessApp = true
	if err != nil && err.Error() == notAccessErr {
		err = nil
		isAccessApp = false
	}
	return
}

func getCloudflareAccessToken(u *url.URL, appInfo *token.AppInfo) (authToken string, err error) {
	authToken, err = token.FetchTokenWithRedirect(u, appInfo, nil)
	return
}
