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

type Image struct{
	filePath string
	img []byte
}

func main() {

	dir := "myDownloadImage"+ time.Now().Format("15_04_05")
	if _,err :=  os.Stat(dir); err != nil {
		os.Mkdir(dir,os.ModeDir)
	}

	photos := Photos{}

	err := getJson("https://jsonplaceholder.typicode.com/photos", &photos)

	if err != nil {
		log.Fatal(err)
	}
   chImg := make(chan Image,len(photos))
   counter := sync.WaitGroup{}
   token := make(chan struct{},20)
   for _, v := range photos {
	   

	   v := v
	   counter.Add(1)	   
      go func ()  {

		defer	counter.Done()

		token <- struct{}{}

		img,err :=  downloadImg(v.ThumbnailURL)
		<- token
		if err != nil {
			 log.Fatal(err)
		}
	 
		
	  
		format,err := decodeImage(img)
		if err != nil {
			log.Fatal(err)
		}
		fileName := (filepath.Join(dir,fmt.Sprintf("%d.%s",v.ID,format)))
		chImg <- Image{filePath: fileName,img: img}
	
	  }()
	   
   }
   go func ()  {

	counter.Wait()
	close(chImg)
	   
   }()

   for  v := range chImg {




	err = saveImage(v.filePath,v.img)
	 
	if err != nil {
		log.Println(err)
	}
	   
   }




}

func saveImage(fileName string,img []byte)  error{
	
    f,err := os.Create(fileName)


	if err != nil {
return err
	}
	defer f.Close()
	 _,err = io.Copy(f,bytes.NewReader(img))

	 if err != nil {
return err
	 }

	 return nil
	
}

func decodeImage(img []byte)   (string,error) {

	_,format,err := image.Decode(bytes.NewReader(img))



	return format ,err
	
}

func downloadImg(url string) ([]byte,error) {

	 res ,err := http.Get(url)

	 if err != nil {
		 return  nil,err
	 }
	defer res.Body.Close()

	body ,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return  nil,err
	}
	return body ,nil
}

func getJson(url string, structType interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return  fmt.Errorf("Get json : %v",err)
	}
	defer res.Body.Close()

	switch v := structType.(type){
	case *Photos : 

	decoder := json.NewDecoder(res.Body)
	results := structType.(*Photos)
	decoder.Decode(results)
	   return nil 
    default : 
    return fmt.Errorf("getJson : Type are not accepted %v",v)

	}
        
	



}