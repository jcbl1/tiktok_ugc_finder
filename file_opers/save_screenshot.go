package fileopers

import (
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func SaveScreenshot(ss []byte) error {
	tail := uuid.NewString()
	tail = strings.Split(tail, "-")[0]
	filename := "screenshot-" + time.Now().Local().Format("20060102150405") + "-" + tail + ".png"
	f, err := os.OpenFile(workingDir+"/"+filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(ss)
	if err != nil {
		return err
	}
	return nil
}
