package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Photos []struct {
	AlbumID      int    `json:"albumId"`
	ID           int    `json:"id"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

type Image struct {
	filePath string
	img      []byte
}

func main() {

	dir := "mypics" + time.Now().Format("15_04_05")
	if _, err := os.Stat(dir); err != nil {
		os.Mkdir(dir, os.ModeDir)

	}

	photos := Photos{}

	if err := getJson("https://jsonplaceholder.typicode.com/photos", &photos); err != nil {
		log.Fatal("Main :", err)
	}

	chImg := make(chan Image)

	token := make(chan struct{}, 2)

	counter := sync.WaitGroup{}
	for _, v := range photos {

		counter.Add(1)

		v := v
		go func() {

			defer counter.Done()

			token <- struct{}{}

			img, err := downloadImage(v.ThumbnailURL)

			<-token
			if err != nil {
				log.Fatal(err)
			}

			format, err := decodeImage(img)
			if err != nil {
				log.Fatal(err)
			}

			filename := filepath.Join(dir, fmt.Sprintf("%d.%s", v.ID, format))
			chImg <- Image{filePath: filename, img: img}

		}()

	}

	go func() {

		counter.Wait()
		close(chImg)

	}()

	for v := range chImg {

		err := saveImage(v.filePath, v.img)

		if err != nil {
			log.Fatal(err)
		}

	}

}

func saveImage(fileName string, img []byte) error {

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(img))

	if err != nil {
		return err
	}

	return nil

}

func decodeImage(img []byte) (string, error) {

	_, f, err := image.Decode(bytes.NewReader(img))

	return f, err

}
func downloadImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

func getJson(url string, stype interface{}) error {

	res, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("GetJson : %v", err)
	}
	defer res.Body.Close()

	switch v := stype.(type) {
	case *Photos:
		decoder := json.NewDecoder(res.Body)
		results := stype.(*Photos)
		decoder.Decode(results)
		return nil
	default:
		return fmt.Errorf("get Json : %v", v)
	}

}
