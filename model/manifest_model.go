package model

type Metadata struct {
	Version  string
	Filetype string
}

type Bundle struct {
	// Name of the bundle
	NA string
	// Version of the bundle
	VE string
	// Timestamp of the bundle
	TS int64
	// Exclude path list
	EX []string
	// Public key string
	PK string
	// Critical flag, if true, any files found not in the manifest will cause a failure
	CR bool
	// Determines if the bundle is signed
	SN bool
	// Root path of the bundle
	RO string
	// Size of the files in the bundle in bytes
	SZ uint64
	// Integrity checksum of the bundle
	CS string
	// Signed hash of the bundle
	SH string
	// List of files in the bundle
	FL Files
}

type File struct {
	// FP is the relative path to the root of the bundle
	FP string
	// FI is the checksum of the file
	FI string
}

type Files []File

type ManifestData struct {
	Metadata Metadata
	Bundle   Bundle
}

type Config struct {
	BundleName     string
	SnapshotName   string
	Version        string
	Critical       bool
	Root           string
	Ignore         []string
	Signed         bool
	PrivateKeyPath string
	PublicKeyPath  string
}
