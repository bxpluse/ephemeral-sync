package ops

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"path/filepath"
)

func DirIdentical(src, dest string) bool {

	srcDirHash := 0
	destDirHash := 0

	err := filepath.Walk(src,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			info.Size()
			srcDirHash = srcDirHash ^ int(hash(info.Name()))
			return nil
		})
	if err != nil {
		return false
	}

	err = filepath.Walk(dest,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			destDirHash = destDirHash ^ int(hash(info.Name()))
			return nil
		})
	if err != nil {
		return false
	}

	if srcDirHash != destDirHash {
		return false
	}

	return true
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func DeleteFile(path string) {
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
}

func DeleteDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func Mkdir(path string) {
	os.Mkdir(path, os.ModeDir)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
