package main

import (
	"errors"
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "CachingS3Reader Suite")
}

var _ = Describe("CachingS3Reader", func() {
	var (
		output []byte
		err    error
	)
	Context("Reading Text", func() {
		BeforeEach(func() {
			r := NewReader("some-bucket-name", "some-special-key", func() (string, error) { return "", errors.New("Error!!!") })
			output, err = ioutil.ReadAll(r)
		})
		It("should just work", func() {
			Expect(err).To(BeNil())
			Expect(string(output)).To(Equal("some-special-key"))
		})
	})
})
