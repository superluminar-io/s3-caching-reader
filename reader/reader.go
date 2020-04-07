package reader

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3CachingReader struct {
	bucketName   string
	key          string
	originReader func() (string, error)
	done         bool
	s3Client     s3iface.S3API
	cacheSeconds int
}

func NewReader(
	bucketName string,
	key string,
	originFunc func() (string, error),
	cacheSeconds int,
	s3Client s3iface.S3API,
) *S3CachingReader {
	return &S3CachingReader{
		bucketName:   bucketName,
		key:          key,
		originReader: originFunc,
		done:         false,
		s3Client:     s3Client,
		cacheSeconds: cacheSeconds,
	}
}

func (r *S3CachingReader) Read(p []byte) (n int, err error) {
	if r.done {
		return 0, io.EOF
	}
	item, err := r.fetchFromS3()
	if err != nil {
		return 0, err
	}
	if item != "" {
		r.done = true
		stringReader := strings.NewReader(item)
		return stringReader.Read(p)
	}

	originItem, err := r.originReader()
	if err != nil {
		return 0, err
	}
	err = r.cacheItem(r.key, originItem)
	if err != nil {
		fmt.Println("failed to write to cache")
	}
	r.done = true
	stringReader := strings.NewReader(originItem)
	return stringReader.Read(p)
}

func (r S3CachingReader) fetchFromS3() (string, error) {
	modifiedSince := modifiedSinceSeconds(r.cacheSeconds)
	resp, err := r.s3Client.GetObject(&s3.GetObjectInput{
		Bucket:          aws.String(r.bucketName),
		IfModifiedSince: &modifiedSince,
		Key:             aws.String(r.key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return "", nil
			case "NotModified":
				return "", nil
			default:
				fmt.Println(aerr.Error())
				return "", aerr
			}
		}
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(body), nil
}

func modifiedSinceSeconds(cacheDurationInSeconds int) time.Time {
	negative := time.Duration(cacheDurationInSeconds*-1) * time.Second
	return time.Now().Add(negative)
}

func (r S3CachingReader) cacheItem(key string, item string) error {
	_, err := r.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(strings.NewReader(item)),
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	originFunc := func() (string, error) {
		return fmt.Sprintf("something from origin at: %s", time.Now().String()), nil
	}
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)
	r := NewReader(
		"s3-caching-reader-test-bucket",
		"my-key-generated-from-some-input",
		originFunc,
		10,
		s3Client,
	)

	all, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf(string(all))
}
