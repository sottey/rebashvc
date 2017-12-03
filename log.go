package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type LogItem struct {
	Name    string
	Message string
}

type LogConcat struct {
	File string
}

func removeDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func setupLogging() {

	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		fmt.Println("Unable to create cache dir!")
		fmt.Println(err)
		os.Exit(1)
	}
	err = os.MkdirAll(logCachePath, 0755)
	if err != nil {
		fmt.Println("Unable to create log dir!")
		fmt.Println(err)
		os.Exit(1)
	}

	removeDirContents(logCachePath)
	go MainLogger(Options.LogPath)
}

func SingleLogger(SingleLogChan chan LogItem, name, logPath string) {

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Unable to create log!")
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	defer func() {
		mainLogConcatChan <- LogConcat{logPath}
	}()

	logger := log.New(file, "", log.Ldate|log.Ltime)
	logger.Println(bold("Task full output: " + name))
	logger.SetFlags(0)

	for {
		select {
		case logObj, ok := <-SingleLogChan:
			if ok {
				logger.Print(logObj.Message)
			} else {
				SingleLogChan = nil
			}
		}
		if SingleLogChan == nil {
			break
		}
	}

}

func MainLogger(logPath string) {

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Unable to create main log!")
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	logger := log.New(file, "", log.Ldate|log.Ltime)

	for {
		select {
		case logObj, ok := <-mainLogChan:
			if ok {
				logger.Print(logObj.Message)
			} else {
				mainLogChan = nil
			}

		case logCmd, ok := <-mainLogConcatChan:
			if ok {
				file.Close()

				out, err := exec.Command("bash", "-c", "cat "+logCmd.File+" >> "+logPath).CombinedOutput()

				if err != nil {
					fmt.Println("Unable to concat logs!")
					fmt.Printf("%s", out)
					fmt.Println(err)
					os.Exit(1)
				}

				file, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					fmt.Println("Unable to create main log!")
					fmt.Println(err)
					os.Exit(1)
				}
				logger = log.New(file, "", log.Ldate|log.Ltime)

				os.Remove(logCmd.File)
			} else {
				mainLogConcatChan = nil
			}
		}
		if mainLogChan == nil && mainLogConcatChan == nil {
			break
		}
	}

	logger.Println(bold("Finished!"))
}