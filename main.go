package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type CachingS3Reader struct {
	bucketName   string
	key          string
	originReader func() (string, error)
	done         bool
}

func NewReader(bucketName, key string, originFunc func() (string, error)) *CachingS3Reader {
	return &CachingS3Reader{
		bucketName:   bucketName,
		key:          key,
		originReader: originFunc,
		done:         false,
	}
}

func (r *CachingS3Reader) Read(p []byte) (n int, err error) {
	if r.done {
		return 0, io.EOF
	}
	item, err := fetchFromCache(r.key)
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
	err = cacheItem(r.key, originItem)
	if err != nil {
		return 0, nil
	}
	stringReader := strings.NewReader(originItem)
	return stringReader.Read(p)
}

func cacheItem(key string, item string) error {
	// TODO
	// - implement write PutObject to S3 with key
	return nil
}

func fetchFromCache(key string) (string, error) {
	// TODO
	// - implement HeadObject from S3 with If-Modified-Since = N minutes
	// - return error? if s3 answers: 304 (not modified)
	return key, nil
}

func main() {
	originFunc := func() (string, error) {
		return "", nil
	}
	r := NewReader("my-cache-bucket", "my-key-generated-from-some-input", originFunc)

	all, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf(string(all))
}
