package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var filename string
	var outputfolder string
	var appname string

	args := os.Args
	for i := range args {
		if os.Args[i] == "-f" {
			filename = os.Args[i+1]
		} else if os.Args[i] == "-o" {
			outputfolder = os.Args[i+1]
		}
	}
	if filename == "" {
		help()
		fmt.Println("\nError: You need to define the IPA file.")
		os.Exit(1)
	}
	if outputfolder == "" {
		help()
		fmt.Println("\nError: You need to define the output folder.")
		os.Exit(1)
	}
	if !fileExists(filename) {
		help()
		fmt.Println("\nError: File " + filename + " not found!")
		os.Exit(1)
	}
	fmt.Println("Extracting file " + filename + " to " + outputfolder)
	_, err := Unzip(filename, outputfolder)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Unzipped:\n" + strings.Join(files, "\n"))

	files, err := ioutil.ReadDir("./" + outputfolder + "/Payload/")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.Contains(f.Name(), ".app") {
			appname = strings.Replace(f.Name(), ".app", "", -1)
		}
	}

	if appname == "" {
		fmt.Println("Error: Cannot find application name!")
		os.Exit(1)
	}

	realpath := "./" + outputfolder + "/Payload/" + appname + ".app/"

	if fileExists(realpath + appname) {
		fmt.Println("Binary file founded: " + realpath + appname)
	}

	result := findStrings(realpath+appname, `([a-zA-Z0-9\-]+):\/\/([a-zA-Z0-9\/\?\.=\-\#]+)`)
	if len(result) > 0 {
		for _, x := range result {
			fmt.Println(x)
		}
	} else {
		fmt.Println("Result: Cannot find any deeplinks hardcoded in the app.")
	}
}

func findStrings(filename string, rgx string) []string {
	file, err := os.Open(filename)
	if err != nil {
		println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 128*128)
	scanner.Buffer(buf, 10*128*128)

	r, _ := regexp.Compile(rgx)
	res := []string{}
	for scanner.Scan() {
		if r.MatchString(scanner.Text()) {
			temp := r.FindAllString(scanner.Text(), -1)
			res = append(res, temp...)
		}
	}
	return res
}

func help() {
	fmt.Println("iOSDeepLinkFinder\nUsage:\n\t-f\tDefine the IPA file.\n\t-o\tDefine output folder.")
}

// fileExists function taken from : https://golangcode.com/check-if-a-file-exists/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Unzip function taken from : https://golangcode.com/unzip-files-in-go/
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}
		filenames = append(filenames, fpath)
		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
