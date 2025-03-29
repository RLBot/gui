package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

type GhRelease struct {
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	HTMLURL         string    `json:"html_url"`
	ID              int       `json:"id"`
	Author          GhAuthor  `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []GhAsset `json:"assets"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

type GhAuthor struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GhAsset struct {
	URL                string    `json:"url"`
	ID                 int       `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           GhAuthor  `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

func ParseReleaseData(data []byte) (GhRelease, error) {
	var release GhRelease
	err := json.Unmarshal(data, &release)
	if err != nil {
		return GhRelease{}, err
	}

	return release, nil
}

// taken from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func extractTar(tr *tar.Reader, dst string) error {
	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

func DownloadExtractArchive(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// body is a .tar.xz file
	xzr, err := xz.NewReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	tr := tar.NewReader(xzr)
	err = extractTar(tr, dest)
	return err
}

func (a *App) GetLatestReleaseData(repo string) (*GhRelease, error) {
	for _, release := range a.latest_release_json {
		if release.repo == repo {
			return &release.content, nil
		}
	}

	latest_release_url := "https://api.github.com/repos/" + repo + "/releases/latest"

	resp, err := http.Get(latest_release_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content, err := ParseReleaseData(body)
	if err != nil {
		return nil, err
	}

	a.latest_release_json = append(a.latest_release_json, RawReleaseInfo{repo, content})

	return &a.latest_release_json[len(a.latest_release_json)-1].content, nil
}

func (a *App) UpdateBotpack(repo string, installPath string, currentTag string) (string, error) {
	latest_release, err := a.GetLatestReleaseData(repo)
	if err != nil {
		return "", err
	}

	newest_tag := latest_release.TagName
	if newest_tag == currentTag {
		return "", errors.New("already up to date")
	}

	current_version := strings.Split(currentTag, "-")[1]
	current_version_num, err := strconv.Atoi(current_version)
	if err != nil {
		return "", err
	}

	newest_version := strings.Split(newest_tag, "-")[1]
	newest_version_num, err := strconv.Atoi(newest_version)
	if err != nil {
		return "", err
	}

	var file_name string
	// todo: check platform (x86_64, etc)
	if runtime.GOOS == "windows" {
		file_name = "patch_x86_64-windows.bobdiff"
	} else {
		file_name = "patch_x86_64-linux.bobdiff"
	}

	var lastest_download_url string
	for _, asset := range latest_release.Assets {
		if asset.Name == file_name {
			lastest_download_url = asset.BrowserDownloadURL
			break
		}
	}

	for i := current_version_num + 1; i <= newest_version_num; i++ {
		tag_i := strings.Replace(currentTag, current_version, strconv.Itoa(i), 1)
		download_url := strings.Replace(lastest_download_url, newest_tag, tag_i, 1)

		resp, err := http.Get(download_url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		xzr, err := xz.NewReader(bytes.NewReader(body))
		if err != nil {
			return "", err
		}

		bytes, err := io.ReadAll(xzr)
		if err != nil {
			return "", err
		}

		files, err := os.ReadDir(installPath)
		if err != nil {
			return "", err
		}

		var dir string
		for _, file := range files {
			if file.IsDir() {
				dir = file.Name()
				break
			}
		}
		if dir == "" {
			return "", errors.New("no directory found")
		}

		err = diff_apply(filepath.Join(installPath, dir), bytes)
		if err != nil {
			return "", err
		}
	}

	return latest_release.TagName, nil
}
