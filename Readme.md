# S3 Caching Reader

![](https://github.com/superluminar-io/s3-caching-reader/workflows/Go/badge.svg)

S3 Caching Reader is a Go module that allows to use S3 as caching backend.

## Installation

```bash
go get github.com/superluminar-io/s3-caching-reader
```

## Usage

```go
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	rdr "github.com/superluminar-io/s3-caching-reader/reader"
)

func main() {
	originFunc := func() (string, error) {
		return fmt.Sprintf("something from origin at: %s", time.Now().String()), nil
	}
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)
	r := rdr.NewReader(
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
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)
