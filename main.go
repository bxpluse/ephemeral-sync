package main

import (
	"./color"
	"./ops"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	srcRoot := "D:/Files"
	destRoot := "E:/Files"

	synPath := "Dir"

	epheSync(srcRoot, destRoot, synPath, 0)
}

func epheSync(srcRoot, destRoot, syncPath string, level int) {

	// Protect D:/ drive from being modified
	if string(srcRoot[0]) != "D" {
		return
	}

	absolutePath := srcRoot + "/" + syncPath
	padding := createSpaces(level)
	files, err := ioutil.ReadDir(absolutePath)

	if err != nil {
		log.Fatal(err)
	}

	cleanUp(srcRoot, destRoot, syncPath, level+1)
	counter := 1

	for _, f := range files {
		fileName := f.Name()
		if f.IsDir() {
			syncDir(srcRoot, destRoot, syncPath, fileName, level)
		} else {
			// Periodically show progress for large dirs
			if counter%1000 == 0 {
				fmt.Println(fmt.Sprintf("%s%s <- Checking dir ... %d", padding, syncPath, counter))
			}
			syncFile(srcRoot, destRoot, syncPath, fileName, level)
		}
		counter += 1
	}
}

func syncDir(srcRoot, destRoot, syncPath, fileName string, level int) {
	newSyncPath := syncPath + "/" + fileName
	newSrcAbsPath := srcRoot + "/" + newSyncPath
	newDestAbsPath := destRoot + "/" + newSyncPath
	padding := createSpaces(level)

	if !ops.Exists(newDestAbsPath) {
		// Create directory in backup drive path
		ops.Mkdir(newDestAbsPath)
	}

	if !ops.DirIdentical(newSrcAbsPath, newDestAbsPath) {
		printAction(padding, fileName, newSyncPath, color.Green, "Syncing dir")
		epheSync(srcRoot, destRoot, newSyncPath, level+1)
		// cleanUp(srcRoot, destRoot, newSyncPath, level + 1)
	} else {
		printAction(padding, fileName, newSyncPath, color.Default, "No diff found")
	}
}

func syncFile(srcRoot, destRoot, syncPath, fileName string, level int) {
	newSyncPath := syncPath + "/" + fileName
	newSrcAbsPath := srcRoot + "/" + newSyncPath
	newDestAbsPath := destRoot + "/" + newSyncPath
	padding := createSpaces(level)
	hashDest, err := md5sum(newDestAbsPath)

	// If error in hashing file, then the dest file does not exist
	if err != nil {
		// Copy file to backup drive path
		printAction(padding, fileName, newSyncPath, color.Green, "Syncing file")
		ops.CopyFile(newSrcAbsPath, newDestAbsPath)
	} else {
		// Copy file if hash mismatches
		hashSrc, _ := md5sum(newSrcAbsPath)
		if hashDest != hashSrc {
			printAction(padding, fileName, newSyncPath, color.Green, "Syncing file")
			ops.CopyFile(newSrcAbsPath, newDestAbsPath)
		}
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

		if !ops.Exists(newSrcAbsPath) {
			if f.IsDir() {
				// Delete all files in the directory, then the dir itself
				printAction(padding, fileName, newSyncPath, color.Red, "Deleting dir")
				ops.DeleteDir(newDestAbsPath)
				ops.DeleteFile(newDestAbsPath)
			} else {
				// Delete the file
				printAction(padding, fileName, newSyncPath, color.Red, "Deleting file")
				ops.DeleteFile(newDestAbsPath)
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

func printAction(padding, name, path, color, action string) {
	fmt.Println(color, fmt.Sprintf("%s%s [%s]  <- %s", padding, name, path, action))
}
