package manifest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"manifest/model"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/IGLOU-EU/go-wildcard/v2"
	"github.com/hashicorp/go-version"
)

type Manifest struct {
	manifest *model.ManifestData
	config   model.Config
}

func Create(config model.Config) (Manifest, error) {
	ver, err := version.NewVersion(config.Version)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		&model.ManifestData{
			Metadata: model.Metadata{
				Version:  "1.0",
				Filetype: "com.orphoros.file.manifest",
			},
			Bundle: model.Bundle{
				NA: config.BundleName,
				VE: ver.String(),
				TS: time.Now().Unix(),
				EX: config.Ignore,
				CR: config.Critical,
				RO: config.Root,
				SN: config.Signed,
				SZ: 0,
				CS: "",
				FL: model.Files{},
			},
		},
		config,
	}, nil
}

func (m Manifest) GetBundle() model.Bundle {
	return m.manifest.Bundle
}

func FromFile(path string) (Manifest, error) {
	var manifest model.ManifestData
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&manifest)
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		&manifest,
		model.Config{},
	}, nil
}

func (m Manifest) ReadConfig(path string) {
	file := filepath.Join(path, "Manifestfile")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if os.IsNotExist(err) {
			log.Fatal("Config file not found")
		} else {
			log.Fatal(err)
		}
	}
	_, err := toml.DecodeFile(path, &m.config)
	if err != nil {
		log.Fatal(err)
	}
}

func (m Manifest) AddFiles() {
	err := filepath.Walk(m.manifest.Bundle.RO,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				for _, ignore := range m.manifest.Bundle.EX {
					if wildcard.Match(ignore, path) {
						return nil
					}
				}
				m.appendFile(path, uint64(info.Size()))
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func (m Manifest) appendFile(path string, size uint64) {
	m.manifest.Bundle.FL = append(m.manifest.Bundle.FL, model.File{
		FP: path,
		FI: getChecksumFile(path),
	})

	m.manifest.Bundle.SZ += size
}

func getChecksumFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (m Manifest) calculateManifestChecksum() {
	m.manifest.Bundle.CS = ""
	data, err := json.Marshal(m.manifest)
	if err != nil {
		log.Fatal(err)
	}
	sha256 := sha256.Sum256(data)
	m.manifest.Bundle.CS = base64.StdEncoding.EncodeToString(sha256[:])
}

func (m Manifest) WriteToFile() error {
	m.calculateManifestChecksum()
	jsonData, err := json.Marshal(m.manifest)
	if err != nil {
		return err
	}

	// check if config is set
	if m.config.SnapshotName == "" {
		return fmt.Errorf("snapshot name not set in config")
	}

	file, err := os.Create(m.config.SnapshotName + ".manifest")
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

func (m Manifest) Check() error {
	if err := m.checkHash(); err != nil {
		return err
	}

	if m.manifest.Bundle.CR {
		return m.checkFileIntegrity()
	}

	return nil
}

func (m Manifest) checkFileIntegrity() error {
	// check if all files are present
	expectedFiles := len(m.manifest.Bundle.FL)
	actualFiles := 0

	err := filepath.Walk(m.manifest.Bundle.RO,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				for _, ignore := range m.manifest.Bundle.EX {
					if wildcard.Match(ignore, path) {
						return nil
					}
				}
				actualFiles++
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	if expectedFiles != actualFiles {
		return fmt.Errorf("file count mismatch")
	}

	return nil
}

func (m Manifest) checkHash() error {
	manifestHash := m.manifest.Bundle.CS
	m.calculateManifestChecksum()
	if manifestHash != m.manifest.Bundle.CS {
		return fmt.Errorf("manifest integrity check failed")
	}

	for _, file := range m.manifest.Bundle.FL {
		if file.FI != getChecksumFile(file.FP) && !strings.Contains(file.FP, ".manifest") {
			return fmt.Errorf("file integrity check failed for %s", file.FP)
		}
	}

	return nil
}
