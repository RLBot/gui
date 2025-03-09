//go:build !production

package main

func (a *App) CheckForNewRelease(repo string, tag string) (bool, error) {
	println("Skipping CheckForNewRelease because we are in dev mode")

	return false, nil
}
