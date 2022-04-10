package main

import (
	"archive/zip"
	"csv_zipper/qrsvg"
	"encoding/csv"
	"fmt"
	"github.com/ajstarks/svgo"
	"github.com/boombuler/barcode/qr"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// generate batch promos
// build csv with promos
// build qrs
// zip qrs

const baseURL = "https://www.google.com/search?q="

func main() {
	year, month, day := time.Now().Date()
	directoryName := fmt.Sprintf("%v_%v_%v_%v_%v", year, month, day, time.Now().Hour(), time.Now().Minute())
	fmt.Println(directoryName)
	createDirectory(directoryName)
	createDirectory(directoryName + "/qrs")

	qrCodes := generateCodes(directoryName)
	generateCsv(qrCodes, directoryName)

	if err := zipSource(directoryName, directoryName+".zip"); err != nil {
		log.Fatal(err)
	}
}

func generateCodes(directoryName string) [][]string {
	var slice [][]string

	for i := 0; i <= 3000; i++ {
		code := randomString(6)
		if i == 0 {
			codeUrls := []string{"code", "url"}
			slice = append(slice, codeUrls)
		} else {
			codeUrls := []string{code, baseURL + code}
			slice = append(slice, codeUrls)
			svgGenerate(directoryName, code, baseURL+code)
		}
	}

	return slice
}

func createDirectory(name string) {
	err := os.Mkdir(name, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func generateCsv(records [][]string, directoryName string) {
	csvPath := fmt.Sprintf("./%s/codes.csv", directoryName)
	f, err := os.Create(csvPath)
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
}

func svgGenerate(directoryName, filename, url string) {
	path := fmt.Sprintf("./%s/qrs/%s", directoryName, filename+".svg")
	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	s := svg.New(f)

	// Create the barcode
	qrCode, _ := qr.Encode(url, qr.M, qr.Auto)

	// Write QR code to SVG
	qs := qrsvg.NewQrSVG(qrCode, 14)
	qs.StartQrSVG(s)
	qs.WriteQrSVG(s)

	s.End()
}

func zipSource(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
