package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v3"
)

var (
	s3BucketName string
	s3KeyPrefix  string
	awsProfile   string
	directUpload bool
)

func init() {
	flag.StringVar(&s3BucketName, "bucket", "", "Required, the name of the S3 bucket. Example: for s3://helmcharts-demo/my-demo you enter helmcharts-demo")
	flag.StringVar(&s3KeyPrefix, "directory", "", "Optional, the path to the directory in the S3 bucket that contains the index file that needs to be restored and the charts. In case this is not the root of the s3 bucket. Example: for s3://helmcharts-demo/my-demo you enter my-demo")
	flag.StringVar(&awsProfile, "profile", "", "Required, the AWS profile configured locally to connect to AWS. Example: default")
	flag.BoolVar(&directUpload, "upload", false, "Upload your new index.yaml directly to S3")
	flag.Parse()

	if len(s3BucketName) == 0 || len(awsProfile) == 0 {
		fmt.Println("Not all required arguments are set:")
		flag.PrintDefaults()
		log.Fatal("Please set all required arguments")
	}

	if len(s3KeyPrefix) > 0 && !strings.HasSuffix(s3KeyPrefix, "/") {
		s3KeyPrefix = s3KeyPrefix + "/"
	}
}

func main() {
	fmt.Println("Starting the restoration of s3://" + s3BucketName + "/" + s3KeyPrefix + "index.yaml")

	config, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithSharedConfigProfile(awsProfile))
	if err != nil {
		log.Fatal("Configuration error: " + err.Error())
	}

	s3Reader := s3.NewFromConfig(config)

	objectsList, err := s3Reader.ListObjects(context.TODO(), &s3.ListObjectsInput{Bucket: &s3BucketName, Prefix: &s3KeyPrefix})
	if err != nil {
		log.Fatal("List bucket content error: " + err.Error())
	}

	index := Index{Entries: make(map[string][]Entry)}
	apiVersions := []string{}

	for _, objectListing := range objectsList.Contents {
		if !strings.HasSuffix(*objectListing.Key, ".tgz") {
			continue
		}

		url := "s3://" + s3BucketName + "/" + *objectListing.Key

		fmt.Println("Parsing information from " + url)

		object, err := s3Reader.GetObject(context.TODO(), &s3.GetObjectInput{Bucket: &s3BucketName, Key: objectListing.Key})
		if err != nil {
			log.Fatal("Get object error: " + err.Error())
		}

		gp, err := gzip.NewReader(object.Body)
		if err != nil {
			log.Fatal("Reading Gzip file: " + err.Error())
		}
		tr := tar.NewReader(gp)
		for {
			h, err := tr.Next()

			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal("Reading Tar file: " + err.Error())
			}

			if strings.HasSuffix(h.Name, "Chart.yaml") {
				entry := Entry{}
				fileContent, err := ioutil.ReadAll(tr)
				if err != nil {
					log.Fatal("Error reading chart.yaml: " + err.Error())
				}

				err = yaml.Unmarshal(fileContent, &entry)
				if err != nil {
					log.Fatal("Error parsing chart.yaml: " + err.Error())
				}

				entry.Urls = append(entry.Urls, url)
				entry.Created = objectListing.LastModified.Format(time.RFC3339Nano)
				entry.Digest, err = digistFromMetadata(object.Metadata)
				if err != nil {
					log.Fatal("Error parsing metadata: " + err.Error())
				}

				index.Entries[entry.Name] = append(index.Entries[entry.Name], entry)
				apiVersions = append(apiVersions, entry.ApiVersion)

				break
			}
		}
	}

	sort.Strings(apiVersions)
	index.ApiVersion = apiVersions[len(apiVersions)-1]
	index.Generated = time.Now().Format(time.RFC3339Nano)

	var data bytes.Buffer

	fmt.Println("Generating new index.yaml")

	yamlEncoder := yaml.NewEncoder(&data)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(&index)
	if err != nil {
		log.Fatal("Error creating index.yaml: " + err.Error())
	}

	if directUpload {
		fmt.Println("Uploading new index.yaml")

		indexKey := s3KeyPrefix + "index.yaml"
		_, err = s3Reader.PutObject(context.TODO(), &s3.PutObjectInput{Bucket: &s3BucketName, Key: &indexKey, Body: bytes.NewReader(data.Bytes())})

		if err != nil {
			log.Fatal("Error writing index.yaml to bucket:" + err.Error())
		}

		fmt.Println("Succesfully restored s3://" + s3BucketName + "/" + s3KeyPrefix + "index.yaml")
	} else {
		err = ioutil.WriteFile("index.yaml", data.Bytes(), 0664)
		if err != nil {
			log.Fatal("Error writing index.yaml to file:" + err.Error())
		}

		fmt.Println("Succesfully generated index.yaml")
	}
}

func digistFromMetadata(metadata map[string]string) (string, error) {
	for key, value := range metadata {
		if key == "chart-digest" {
			return value, nil
		}
	}

	return "", errors.New("No digist found in the metadata")
}
