// Package ugcinfo defines some UGC structures.
package ugcinfo

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// type VideoStats struct {
// 	DiggCount int `json:"digg_count"`
// 	PlayCount int `json:"play_count"`
// }

// UGCInfo is a structure for cared infomation about a UGC.
type UGCInfo struct {
	Name            string    `json:"name"`
	Signature       string    `json:"signature"`
	UniqueID        string    `json:"unique_id"`
	FollowerCount   int       `json:"follower_count"`
	Gender          string    `json:"gender"`
	AP              int       `json:"ap"`
	AI              float32   `json:"ai"`
	Email           []string  `json:"email"`
	LatestVideoTime time.Time `json:"latest_video_time"`
	// VideosStats     []VideoStats
}

// func (u UGCInfo) String() string {
// 	data, _ := json.Marshal(u)
// 	return fmt.Sprintf("%s", data)
// }

// HashtagResultAuthor represents the result of JSON unmarshalling of hashtag results in the field of .author.
type HashtagResultAuthor struct {
	AvatarMedium string `json:"avatarMedium"`
	ID           string `json:"id"`
	Nickname     string `json:"nickname"`
	Signature    string `json:"signature"`
	UniqueID     string `json:"uniqueId"`
}

// HashtagResultAuthorStats represents the result of JSON unmarshalling of hashtag results in the field of .authorStats
type HashtagResultAuthorStats struct {
	DiggCount     int `json:"diggCount"`
	FollowerCount int `json:"followerCount"`
	HeartCount    int `json:"heartCount"`
	VideoCount    int `json:"videoCount"`
}

// type HashtagResultContent struct{
// 	Desc string `json:"desc"`
// }

// type HashtagResultStats struct {
// 	CollectCount int `json:"collectCount"`
// 	CommentCount int `json:"commentCount"`
// 	DiggCount    int `json:"diggCount"`
// 	PlayCount    int `json:"playCount"`
// 	ShareCount   int `json:"shareCount"`
// }

// HashtagResult represents the result of JSON unmarshalling of hashtag results.
type HashtagResult struct {
	Author      HashtagResultAuthor      `json:"author"`
	AuthorStats HashtagResultAuthorStats `json:"authorStats"`
	// Contents []HashtagResultContent
	CreatedTime int    `json:"createdTime"`
	Desc        string `json:"desc"`
}

// HashtagResults represents a slice of HashtagResult
type HashtagResults []HashtagResult

// FromJSON reads from scrapedJSONFile and returns the UGCInfos in it.
func FromJSON(scrapedJSONFile string) ([]UGCInfo, error) {
	var ugcs []UGCInfo
	f, err := os.Open(scrapedJSONFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var hashRess HashtagResults
	json.Unmarshal(buf, &hashRess)

	// Redundancy declusion
	present := make(map[string]bool)
	for _, hashRes := range hashRess {
		if _, ok := present[hashRes.Author.UniqueID]; !ok && hashRes.AuthorStats.FollowerCount >= minFollowerCount && hashRes.AuthorStats.FollowerCount <= maxFollowerCount {
			present[hashRes.Author.UniqueID] = true
			ugcs = append(ugcs, UGCInfo{
				Name:          hashRes.Author.Nickname,
				Signature:     hashRes.Author.Signature,
				UniqueID:      hashRes.Author.UniqueID,
				FollowerCount: hashRes.AuthorStats.FollowerCount,
			})
		}
	}

	return ugcs, nil
}
