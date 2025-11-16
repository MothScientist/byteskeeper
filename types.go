package main

const bufferSize = 64 * 1024 // 128 KB

type File struct {
	Path string
	Name string
}

type FileHash struct {
	File
	Hash  string
	Error error
}

type HashAlgorithm struct {
	code int
	name string
}

var (
	SHA256 = HashAlgorithm{code: 0, name: "sha256"}
	BLAKE2b = HashAlgorithm{code: 1, name: "blake2b"}
)
