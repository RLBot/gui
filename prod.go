//go:build production

package main

func (a *App) CheckForNewRelease(repo string, tag string) (bool, error) {
	latest_release, err := a.GetLatestReleaseData(repo)
	if err != nil {
		return false, err
	}

	return latest_release.TagName != tag, nil
}
