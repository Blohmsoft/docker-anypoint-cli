package main

import (
	"encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var snapshotFlag = flag.Bool("snapshot", false, "set if desired version is snapshot")
var repoFlag = flag.String("repo", "", "Repository host")
var groupIdFlag = flag.String("groupId", "", "Artifact GroupId")
var artifactIdFlag = flag.String("artifactId", "", "ArtifactId")
var versionFlag = flag.String("version", "", "Artifact version")
var filenameFlag = flag.String("filename", "", "Artifact filename (without extension, not used for snapshots)")
var extensionFlag = flag.String("extension", "", "Artifact extension (not used for snapshots)")

var userFlag = flag.String("user", "", "username to artifact repository")
var passFlag = flag.String("pass", "", "password to artifact repository")

type snapshotVersions struct {
	GroupId    string            `xml:"groupId"`
	ArtifactId string            `xml:"artifactId"`
	Version    string            `xml:"version"`
	Snapshots  []snapshotVersion `xml:"versioning>snapshotVersions>snapshotVersion"`
}

type snapshotVersion struct {
	Extension  string `xml:"extension"`
	Value      string `xml:"value"`
	Updated    string `xml:"updated"`
	Classifier string `xml:"classifier"`
}

func addAuth(request *http.Request) {
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(*userFlag+":"+*passFlag))
	request.Header.Add("Authorization", auth)
}

func main() {
	flag.Parse()

	basePath := fmt.Sprintf("%s/%s/%s/%s", *repoFlag, strings.Replace(*groupIdFlag, ".", "/", -1), *artifactIdFlag, *versionFlag)

	if *snapshotFlag {
		downloadSnapshots(basePath, *artifactIdFlag, *versionFlag)
	} else {
		downloadArtifact(basePath, *artifactIdFlag, *versionFlag, *filenameFlag, *extensionFlag)
	}
}

func downloadArtifact(basePath string, artifactId string, version string, filename string, extension string) {
	url := fmt.Sprintf("%s/%s.%s", basePath, filename, extension)
	dest := fmt.Sprintf("%s-%s.%s", artifactId, version, extension)

	log.Printf("Downloading: %s\n", url)
	log.Printf("Writing to: %s", dest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	addAuth(req)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(dest)
	_, _ = io.Copy(out, resp.Body)
}

func downloadSnapshots(basePath string, artifactId string, version string) {
	req, _ := http.NewRequest("GET", basePath+"/maven-metadata.xml", nil)

	addAuth(req)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var snapshotVersions snapshotVersions
	err = xml.NewDecoder(resp.Body).Decode(&snapshotVersions)
	if err != nil {
		fmt.Printf("error when parsing response: %v", err)
	}

	for _, s := range snapshotVersions.Snapshots {
		var url string
		if s.Classifier != "" {
			url = fmt.Sprintf("%s/%s-%s-%s.%s", basePath, artifactId, s.Value, s.Classifier, s.Extension)
		} else {
			url = fmt.Sprintf("%s/%s-%s.%s", basePath, artifactId, s.Value, s.Extension)
		}

		downloadSnapshot(client, url, artifactId, version, s.Extension)
	}
}

func downloadSnapshot(client http.Client, url string, artifactId string, version string, extension string) {
	dest := fmt.Sprintf("%s-%s.%s", artifactId, version, extension)

	log.Printf("Downloading: %s\n", url)
	log.Printf("Writing to: %s", dest)

	fileReq, _ := http.NewRequest("GET", url, nil)
	addAuth(fileReq)
	resp, _ := client.Do(fileReq)
	out, _ := os.Create(dest)
	_, _ = io.Copy(out, resp.Body)
}
