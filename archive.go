package main

import (
	"errors"
	"os"
	"path"
	"strings"
)

func loadArchive(dirPath string) (err error) {
	dirContents, err := os.ReadDir(dirPath)
	if err != nil {
		return errors.New("Could not read archive at: " + dirPath + ": " + err.Error())
	}
	for _, file := range dirContents {
		filePath := path.Join(dirPath, file.Name())
		if file.IsDir() {
			if file.Name() == "users" {
				loadUserArchive(filePath)
			} else {
				loadArchive(filePath)
			}
		} else {
			if strings.HasSuffix(file.Name(), "txt") {
				masterVoice.inputTextFromFile(filePath)
			}
		}
	}

	return
}

//loads the user archive.
//TODO: implement user black/whitelist ignore logic here if that feature ever makes it in
func loadUserArchive(dirPath string) error {
	dirContents, err := os.ReadDir(dirPath)
	if err != nil {
		return errors.New("Could not read user archive at: " + dirPath + ": " + err.Error())
	}

	for _, file := range dirContents {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "txt") {
			filePath := path.Join(dirPath, file.Name())
			masterVoice.inputTextFromFile(filePath)
		}
	}

	return nil
}

//writes a new string to the user's archive at "archive/users/*.txt". creates a new file if required
func writeStringToUserArchive(userID string, s string) error {
	userFile, err := os.OpenFile("archive/users/"+userID+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("Error logging message: " + err.Error())
	}
	defer userFile.Close()

	userFile.WriteString(s + "\n")

	return nil
}
