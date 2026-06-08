package msbimport

import (
	"os"
	"time"
)

func SetImageTime(f string, t time.Time) error {
	return os.Chtimes(f, t, t)
	// asciiTime := t.Format("2006:01:02 15:04:05")
	// _, err := exec.Command("exiv2", "-M", "set Exif.Image.DateTime "+asciiTime, f).CombinedOutput()
	// if err != nil {
	// 	return err
	// }
	// return nil
}
