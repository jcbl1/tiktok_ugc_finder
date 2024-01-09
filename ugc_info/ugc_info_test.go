package ugcinfo

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestFromJSON(t *testing.T) {
	if ugcs, err := FromJSON("/home/jcbl1/tiktok_hashtag_data/faceyoga/posts_fewer.json"); err != nil {
		t.Error(err)
	} else {
		if len(ugcs) != 2 {
			t.Error(errors.New("len not 2"))
		}
		jsonOutput(t, ugcs)
	}

	uuid_ := uuid.NewString()
	if _, err := FromJSON(uuid_); err != nil {
		t.Logf("Error when input wrong filename: %s", err)
	}

	//test decoder
	t.Fail()
}

func jsonOutput(t *testing.T, s interface{}) {
	buf, err := json.Marshal(s)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s\n", buf)
}
