package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/google/go-github/github"
	version "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	pb "gopkg.in/cheggaaa/pb.v1"
)

func getVersionsFromGithub() (versions []*version.Version, err error) {
	gh := github.NewClient(nil)
	ctx := context.Background()
	tags, _, err := gh.Repositories.ListTags(ctx, "gumieri", "ecsctl", &github.ListOptions{})
	if err != nil {
		return
	}

	var versionsRaw []string
	for _, tag := range tags {
		versionsRaw = append(versionsRaw, *tag.Name)
	}

	versions = make([]*version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		versions[i], err = version.NewVersion(raw)
		if err != nil {
			return
		}
	}

	sort.Sort(version.Collection(versions))
	return
}

func uname() (uname string, err error) {
	s := strings.Title(runtime.GOOS)

	var m string
	switch runtime.GOARCH {
	case "386":
		m = "i386"
	case "amd64":
		m = "x86_64"
	case "arm":
		err = errors.New("Unable to identify the ARM architecture")
		return
	}

	uname = "ecsctl-" + s + "-" + m

	if runtime.GOOS == "windows" {
		uname = uname + ".exe"
	}

	return
}

func upgradeRun(cmd *cobra.Command, args []string) {
	available, err := getVersionsFromGithub()
	t.Must(err)
	latest := available[len(available)-1]

	current, err := version.NewVersion(VERSION)
	t.Must(err)

	if !current.LessThan(latest) {
		t.Infoln("You are using the latest version")
		return
	}

	t.Infof("There's a new version available. (current: %s - available: %s)\n", current, latest)
	if !yes && !t.Confirm("Do you want to upgrade?") {
		return
	}

	selfPath, err := os.Executable()
	t.Must(err)

	selfDir := filepath.Dir(selfPath)

	actualFile, err := os.Open(selfPath)
	t.Must(err)

	fileStat, err := actualFile.Stat()
	t.Must(err)

	newFileName := "temp_" + filepath.Base(selfPath)
	newFilePath := filepath.Join(selfDir, newFileName)
	newFile, err := os.Create(newFilePath)
	t.Must(err)
	defer os.Remove(newFilePath)
	newFile.Chmod(fileStat.Mode())

	u, err := uname()
	t.Must(err)

	url := "https://github.com/gumieri/ecsctl/releases/download/v" + latest.String() + "/" + u

	request, err := http.NewRequest("GET", url, nil)
	t.Must(err)

	client := &http.Client{}
	response, err := client.Do(request)
	if response.StatusCode != 200 {
		err = fmt.Errorf("failed to download binary from GitHub. HTTP Status: %d", response.StatusCode)
	}
	t.Must(err)
	defer response.Body.Close()

	var proxyBody io.ReadCloser
	if quiet {
		_, err = io.Copy(newFile, response.Body)
		t.Must(err)
	} else {
		bar := pb.New(int(response.ContentLength)).SetUnits(pb.U_BYTES)
		proxyBody = bar.NewProxyReader(response.Body)

		bar.Start()
		_, err = io.Copy(newFile, proxyBody)
		t.Must(err)
		bar.Finish()
	}

	t.Must(os.Rename(newFilePath, selfPath))
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the ecsctl binary to the latest stable release",
	Run:   upgradeRun,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	flags := upgradeCmd.Flags()
	flags.BoolVarP(&yes, "yes", "y", false, yesSpec)
}
