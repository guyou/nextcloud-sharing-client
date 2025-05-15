package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	openapi "github.com/nextcloud/api-sdk"
	"github.com/studio-b12/gowebdav"
)

type MyRoundTripper struct{}

func (t MyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	// Do work before the request is sent
	log.Printf("%+v\n", req.URL)

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Do work after the response is received

	return resp, err
}

func UploadFileWebDAV(nextcloudURL, username, password, localFilePath, remotePath string) error {
	remoteURL, err := url.Parse(nextcloudURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}
	remoteURL = remoteURL.JoinPath("remote.php/dav/files")

	c := gowebdav.NewClient(remoteURL.String(), username, password)
	c.Connect()

	file, _ := os.Open(localFilePath)
	defer file.Close()

	return c.WriteStream(filepath.Join(username, remotePath), file, 0644)
}

func createPublicShare(apiClient *openapi.APIClient, path string) (*openapi.FilesSharingShare, error) {
	oCSAPIRequest := "oCSAPIRequest_example" // string |  (default to "true")
	shareType := int64(3)
	resp, r, err := apiClient.FilesSharingShareapiApi.FilesSharingShareapiCreateShare(context.Background()).OCSAPIRequest(oCSAPIRequest).Path(path).ShareType(shareType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesSharingShareapiApi.FilesSharingShareapiCreateShare``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	return &resp.Ocs.Data, nil
}

func getPublicShare(apiClient *openapi.APIClient, path string) (*openapi.FilesSharingShare, error) {
	oCSAPIRequest := "oCSAPIRequest_example" // string |  (default to "true")
	resp, r, err := apiClient.FilesSharingShareapiApi.FilesSharingShareapiGetShares(context.Background()).OCSAPIRequest(oCSAPIRequest).Path(path).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilesSharingShareapiApi.FilesSharingShareapiGetShares``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	for _, share := range resp.Ocs.Data {
		if share.ShareType == 3 {
			return &share, nil
		}
	}
	return nil, nil
}

func buildDownladUrl(share *openapi.FilesSharingShare) openapi.NullableString {
	url := openapi.NullableString{}
	filename := filepath.Base(share.Path)
		public := *share.Url.Get() + "/download/" + filename
		url.Set(&public)
	return url
}

func main() {
	baseUrl := flag.String("baseurl", "", "Base URL for Nextcloud server")
	username := flag.String("username", "", "")
	password := flag.String("password", "", "")
	remoteFolder := flag.String("path", "Share", "Remote path")
	upload := flag.Bool("upload", true, "")

	flag.Parse()

	parsedBaseUrl, err := url.Parse(*baseUrl)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	//ctx := context.Background()
	cfg := &openapi.Configuration{
		Scheme: parsedBaseUrl.Scheme,
		Host:   parsedBaseUrl.Host,
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(*username+":"+*password))),
		},
		Servers: openapi.ServerConfigurations{
			{
				URL: *baseUrl,
			},
		},
	}
	client := openapi.NewAPIClient(cfg)

	// Read the latest arguments (after flags)
	latestArgs := flag.Args() // This will give you the non-flag arguments
	if len(latestArgs) > 0 {
		for _, localPath := range latestArgs {
			log.Printf("Processing %s", localPath)
			basename := filepath.Base(localPath)
			remotePath := filepath.Join(*remoteFolder, basename)

			if *upload {
				err = UploadFileWebDAV(*baseUrl, *username, *password, localPath, remotePath)
				if err != nil {
					log.Fatal("Failed to upload:", err)
				}
			}
			share, err := getPublicShare(client, remotePath)
			if err != nil {
				log.Fatal("Failed to retrieve Public Share: ", err)
			}
			if share == nil {
				log.Printf("Creating share for %s\n", remotePath)
				share, err = createPublicShare(client, remotePath)
				if err != nil {
					log.Fatal("Failed to create share: ", err)
				}
			}
			log.Printf("Share: %+v\n", share)
			url := buildDownladUrl(share)
			if url.IsSet() {
				fmt.Fprintf(os.Stdout, "Share URL: %s\n", *url.Get())
			} else {
				fmt.Fprint(os.Stderr, "Failed to find url")
			}
		}
	} else {
		fmt.Println("No additional arguments provided.")
	}

}
