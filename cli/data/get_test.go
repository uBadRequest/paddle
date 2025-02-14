package data

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
)

func TestFilterObjects(t *testing.T) {
	var (
		key1   = "path/file1.csv"
		key2   = "path/file2.csv"
		key3   = "path/folder/file3.csv"
		obj1   = &s3.Object{Key: &key1}
		obj2   = &s3.Object{Key: &key2}
		obj3   = &s3.Object{Key: &key3}
		keys   = []string{"file1.csv", "file2.csv", "folder/file3.csv"}
		s3Path = S3Path{bucket: "bucket", path: "path/"}
	)

	result, err := filterObjects(s3Path, []*s3.Object{obj1, obj2, obj3}, keys)
	if err != nil {
		t.Errorf("It should filter objects properly, but %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Failed to filter keys got: %v, want: 3", len(result))
	}
}

func TestFilterObjectsWithNoKeys(t *testing.T) {
	var (
		key    = "path/file.csv"
		obj    = &s3.Object{Key: &key}
		s3Path = S3Path{bucket: "bucket", path: "path/"}
	)

	result, err := filterObjects(s3Path, []*s3.Object{obj}, []string{})
	if err != nil {
		t.Errorf("It should filter objects properly, but %v", err)
	}

	length := len(result)
	if length != 1 {
		t.Errorf("It should return all objects, but got: %v, want: 1.", length)
	}
}

func TestFilterObjectsUsingNonExistentKeys(t *testing.T) {
	var (
		key   = "path/f1.csv"
		obj    = &s3.Object{Key: &key}
		s3Path = S3Path{bucket: "bucket", path: "path/"}
		keys   = []string{"f2.csv", "f3.csv"}
	)

	result, err := filterObjects(s3Path, []*s3.Object{obj}, keys)
	if result != nil {
		t.Error("It should not return a list of S3 objects")
	}

	if err == nil {
		t.Error("It should return an error")
	}
}
