package utils

import (
	"encoding/json"
	"fmt"
)

func OutputAsJSON(s interface{}) error {
	if buf, err := json.Marshal(s); err != nil {
		return err
	} else {
		fmt.Printf("%s", buf)
	}
	return nil
}
