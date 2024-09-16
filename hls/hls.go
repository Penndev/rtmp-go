package hls

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/penndev/rtmp-go/mpegts"
	"github.com/penndev/rtmp-go/rtmp"
)

var HlsHeader = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:YES
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:<sequence>`

func HandleHls(subTopic func(string) (*rtmp.PubSub, bool)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		param := r.URL.Query()
		topic := param.Get("topic")
		if _, ok := subTopic(topic); ok {
			if c, l, ok := mpegts.HlsLive(topic); ok {
				s := strings.Replace(HlsHeader, "<sequence>", strconv.Itoa(l), 1)
				for _, v := range c {
					s += "\n#EXTINF:" + strconv.Itoa(int(v.Inf/1000)) + "." + strconv.Itoa(int(v.Inf%1000)) + "\n" + v.File
				}
				w.Write([]byte(s))
			} else {
				http.Error(w, "service close", 400)
			}
		} else {
			http.NotFound(w, r)
		}
	}
}
