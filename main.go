package main

import (
	"./color"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {

	srcRoot := "D:/"
	destRoot := "E:/"

	synPath := "SomeDir"

	sync(srcRoot, destRoot, synPath, 0)
}

func sync(srcRoot string, destRoot string, syncPath string, level int) {

	if srcRoot != "D:/" {
		return
	}

	absolutePath := srcRoot + "/" + syncPath
	padding := createSpaces(level)
	newLevel := level + 1
	files, err := ioutil.ReadDir(absolutePath)

	if err != nil {
		log.Fatal(err)
	}

	cleanUp(srcRoot, destRoot, syncPath, newLevel)
	counter := 1
	for _, f := range files {
		fileName := f.Name()
		newSyncPath := syncPath + "/" + fileName
		newSrcAbsPath := srcRoot + "/" + newSyncPath
		newDestAbsPath := destRoot + "/" + newSyncPath

		if f.IsDir() {

			if !exists(newDestAbsPath) {
				// Create directory in backup drive path
				mkdir(newDestAbsPath)
			}

			if !DirIdentical(newSrcAbsPath, newDestAbsPath) {
				printAction(padding, fileName, newSyncPath, color.Default, "Syncing dir")
				sync(srcRoot, destRoot, newSyncPath, level+1)
				cleanUp(srcRoot, destRoot, newSyncPath, newLevel)
			} else {
				printAction(padding, fileName, newSyncPath, color.Default, "No diff found")
			}

		} else {

			// Periodically show progress for large dirs
			if counter%1000 == 0 {
				fmt.Println(fmt.Sprintf("%s%s <- Checking dir ... %d", padding, syncPath, counter))
			}

			hashDest, err := md5sum(newDestAbsPath)

			// If error in hashing file, then the dest file does not exist
			if err != nil {
				// Copy file to backup drive path
				printAction(padding, fileName, newSyncPath, color.Green, "Syncing file")
				copyFile(newSrcAbsPath, newDestAbsPath)
			} else {
				// Copy file if hash mismatches
				hashSrc, _ := md5sum(newSrcAbsPath)
				if hashDest != hashSrc {
					printAction(padding, fileName, newSyncPath, color.Green, "Syncing file")
					copyFile(newSrcAbsPath, newDestAbsPath)
				}
			}
		}
		counter += 1
	}
}

func cleanUp(srcRoot string, destRoot string, syncPath string, level int) {
	absolutePath := destRoot + "/" + syncPath
	padding := createSpaces(level)
	files, _ := ioutil.ReadDir(absolutePath)

	// If file no longer exists in source, it doesn't need to exist in destination
	for _, f := range files {
		fileName := f.Name()
		newSyncPath := syncPath + "/" + fileName
		newSrcAbsPath := srcRoot + "/" + newSyncPath
		newDestAbsPath := destRoot + "/" + newSyncPath

		if !exists(newSrcAbsPath) {
			if f.IsDir() {
				// Delete all files in the directory, then the dir itself
				printAction(padding, fileName, newSyncPath, color.Red, "Deleting dir")
				deleteDir(newDestAbsPath)
				deleteFile(newDestAbsPath)
			} else {
				// Delete the file
				printAction(padding, fileName, newSyncPath, color.Red, "Deleting file")
				deleteFile(newDestAbsPath)
			}
		}
	}
}

func createSpaces(level int) string {
	s := ""
	for i := 0; i < level; i++ {
		s += "    "
	}
	return s
}

func mkdir(path string) {
	os.Mkdir(path, os.ModeDir)
}

func md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), err
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func copyFile(src, dst string) (int64, error) {
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

func deleteFile(path string) {
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
}

func deleteDir(path string) error {
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

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func printAction(padding, name, path, color, action string) {
	fmt.Println(color, fmt.Sprintf("%s%s [%s]  <- %s", padding, name, path, action))
}
