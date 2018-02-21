package test_helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf/dedicated-mysql-utils/testhelpers"
)

type BucketConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	BucketPath      string
	EndpointURL     string
}

func configValue(manifestPath, key string) string {
	return testhelpers.Interpolate(manifestPath, fmt.Sprintf("/properties/service-backup/destinations/type=s3/config/%s", key))
}

func FetchBucketConfig(deploymentName string) BucketConfig {
	manifestPath := testhelpers.DownloadManifest(deploymentName)

	defer func() {
		err := os.Remove(manifestPath)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}()

	return BucketConfig{
		AccessKeyID:     configValue(manifestPath, "access_key_id"),
		SecretAccessKey: configValue(manifestPath, "secret_access_key"),
		BucketName:      configValue(manifestPath, "bucket_name"),
		BucketPath:      path.Join(configValue(manifestPath, "bucket_path"), deploymentName),
		EndpointURL:     configValue(manifestPath, "endpoint_url"),
	}
}

func determineLatestBackup(bucketConfig BucketConfig, awsSession *session.Session) string {
	s3Client := s3.New(awsSession)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketConfig.BucketName),
		Prefix: aws.String(bucketConfig.BucketPath),
	}
	listing, err := s3Client.ListObjects(params)
	Expect(err).ToNot(HaveOccurred())

	if len(listing.Contents) == 0 {
		return ""
	}

	var latestBackup string
	for i := len(listing.Contents) - 1; i >= 0; i-- {
		if strings.HasSuffix(*listing.Contents[i].Key, ".tar.gpg") {
			latestBackup = *listing.Contents[i].Key
			break
		}
	}

	if len(latestBackup) == 0 {
		return ""
	}

	return latestBackup
}

func createAWSSession(bucketConfig BucketConfig) *session.Session {
	creds := credentials.NewStaticCredentials(bucketConfig.AccessKeyID, bucketConfig.SecretAccessKey, "")

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})

	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return s
}

func downloadFile(bucketConfig BucketConfig, awsSession *session.Session, path string) *os.File {
	getParams := &s3.GetObjectInput{
		Bucket: aws.String(bucketConfig.BucketName),
		Key:    aws.String(path),
	}

	downloader := s3manager.NewDownloader(awsSession)
	backupFile, err := ioutil.TempFile("", "mysql-backup")
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	_, err = downloader.Download(backupFile, getParams)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	return backupFile
}

func DownloadBackup(deploymentName string) (BucketConfig, *session.Session, string) {
	bucketConfig := FetchBucketConfig(deploymentName)
	awsSession := createAWSSession(bucketConfig)
	latestBackup := ""
	EventuallyWithOffset(1, func() string {
		latestBackup = determineLatestBackup(bucketConfig, awsSession)
		return latestBackup
	}, "5m", "30s").ShouldNot(BeEmpty(), fmt.Sprintf("Could not download backup from in s3://%s/%s",
		bucketConfig.BucketName,
		bucketConfig.BucketPath))

	backupPath := downloadFile(bucketConfig, awsSession, latestBackup)
	return bucketConfig, awsSession, backupPath.Name()
}

func UploadBackupFile(deploymentName, downloadedFile, targetPath string) {
	uploadPath := fmt.Sprintf("mysql:%s", targetPath)
	args := []string{
		"-d",
		deploymentName,
		"scp",
		downloadedFile,
		uploadPath,
	}
	command := exec.Command(BoshPath, args...)
	command.Stdout = ginkgo.GinkgoWriter
	command.Stderr = ginkgo.GinkgoWriter
	err := command.Run()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

func RestoreBackup(deploymentName, decryptionKey, backupPath string) {
	cmdString := `sudo /var/vcap/packages/mysql-restore/bin/restore --encryption-key %s --restore-file %s`

	ManageInstanceProcesses(deploymentName, "stop", "mysql")
	remoteCommand := fmt.Sprintf(cmdString, decryptionKey, backupPath)
	command := exec.Command(BoshPath, "-d", deploymentName, "ssh", "mysql/0", "-c", remoteCommand)

	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	session.Wait(5 * time.Minute)
	ExpectWithOffset(1, session.ExitCode()).To(Equal(0))

	ManageInstanceProcesses(deploymentName, "start", "mysql")
}
