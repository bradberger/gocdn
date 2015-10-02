package gocdn

import (
    "io/ioutil"
    "os"
    "path"
    "log"
)

func cacheFile(fileName string, data []byte) (err error){

	fileName = path.Clean(fileName)
	dir := path.Dir(fileName)

	if err = os.MkdirAll(dir, os.FileMode(0775)); err != nil {
        log.Printf("Could not create directory: %s", dir)
		return
	}

	if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
        log.Printf("Could not write file: %s", dir)
		return
	}

    return

}
