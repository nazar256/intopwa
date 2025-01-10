package cloudflare

import (
	"fmt"
	"github.com/syumai/workers/cloudflare"
	"io"
)

type r2 struct {
	bucket *cloudflare.R2Bucket
}

func NewBucket(bucket *cloudflare.R2Bucket) *r2 {
	return &r2{
		bucket: bucket,
	}
}

func (r *r2) Get(path string) (io.Reader, error) {
	obj, err := r.bucket.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read from R2 path: %w", err)
	}

	return obj.Body, nil
}

func (r *r2) Put(path string, data io.ReadCloser) error {
	_, err := r.bucket.Put(path, data, nil)

	return err
}
