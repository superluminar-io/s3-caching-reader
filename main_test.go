package cache

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type S3ClientMock struct {
	mock.Mock
	s3iface.S3API
}

func (c S3ClientMock) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	arg := c.Called(input)
	if out, ok := arg.Get(0).(*s3.PutObjectOutput); ok {
		return out, nil
	}
	return nil, arg.Error(1)
}

func (c S3ClientMock) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	arg := c.Called(input)
	if out, ok := arg.Get(0).(s3.GetObjectOutput); ok {
		return &out, nil
	}
	return nil, arg.Error(1)
}

func TestCachingS3Reader(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "S3CachingReader Suite")
}

var _ = Describe("S3CachingReader", func() {
	var (
		bucketName   string
		key          string
		originReader func() (string, error)
		done         bool
		cacheSeconds int
		s3Client     S3ClientMock
		output       []byte
		err          error
		aReader      = func() *S3CachingReader {
			return &S3CachingReader{
				bucketName:   bucketName,
				key:          key,
				originReader: originReader,
				done:         done,
				s3Client:     s3Client,
				cacheSeconds: cacheSeconds,
			}
		}
	)

	BeforeEach(func() {
		bucketName = "some-default-bucket"
		key = "some-default-key"
		originReader = func() (string, error) { return "some-origin-value", nil }
		done = false
		cacheSeconds = 1
		s3Client = S3ClientMock{}
	})

	Context("Reading Text", func() {
		It("returns without an error when done", func() {
			r := aReader()
			r.done = true
			output, err = ioutil.ReadAll(r)
			Expect(err).To(BeNil())
			Expect(output).To(BeEmpty())
		})

		It("returns the item fetched from S3", func() {
			content := []byte("some content")
			getObjectOutput := s3.GetObjectOutput{}
			getObjectOutput.SetBody(ioutil.NopCloser(bytes.NewReader(content)))
			s3Client.On("GetObject", mock.Anything).Return(getObjectOutput, nil)
			s3Client.On("PutObject", mock.Anything).Return(nil, nil)

			r := aReader()
			output, err = ioutil.ReadAll(r)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(content))
		})
	})
})
