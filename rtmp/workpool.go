package rtmp

import (
	"fmt"
	"log"
	"rtmp-go/av"
	"time"
)

// 存储视频元数据
type metaPack struct {
	meta  []byte
	audit []byte
	video []byte
}

// 播放端列表
type stream struct {
	client map[chan Pack]bool
	meta   metaPack
}

// 第 1 bit 设置 onMetaData
// 第 2 bit 设置 Audit
// 第 3 bit 设置 Video
//0000 0111
func (s *stream) setMeta(pk Pack, readyIng *int) {
	if pk.MessageTypeID == 18 || pk.MessageTypeID == 15 {
		s.meta.meta = pk.PayLoad[16:]
		pk.PayLoad = pk.PayLoad[16:]
		*readyIng |= 1 // onMetaData
		// return
	}
	if pk.MessageTypeID == 8 {
		s.meta.audit = pk.PayLoad
		*readyIng |= 2 // onAuditInit
		// return
	}
	if pk.MessageTypeID == 9 {
		s.meta.video = pk.PayLoad
		*readyIng |= 4 // onVideoInit
		// return
	}
	s.setPack(pk)
}

func (s *stream) setPack(pk Pack) {
	for c := range s.client {
		c <- pk
	}
}

type listener func(string) chan Pack

type App struct {
	Gloab map[string]listener
	List  map[string]*stream
}

func (a *App) addGloab(k string, f listener) {
	a.Gloab[k] = f
}

func (a *App) addPublish(appName string, streamName string) *stream {
	pool := appName + "_" + streamName
	app, ok := a.List[pool]
	if !ok {
		a.List[pool] = &stream{
			client: make(map[chan Pack]bool),
			meta:   metaPack{},
		}
		app = a.List[pool]
	}
	// 初始化播放段
	for _, bl := range a.Gloab {
		listen := bl(pool)
		app.client[listen] = true
	}
	return a.List[pool]
}

func (a *App) addPlay(appName string, streamName string, client chan Pack) bool {
	pool := appName + "_" + streamName
	app, ok := a.List[pool]
	if !ok {
		return false
	}
	app.client[client] = true
	return true
}

func (a *App) getMeta(appName string, streamName string) metaPack {
	pool := appName + "_" + streamName
	app, ok := a.List[pool]
	if !ok {
		return metaPack{}
	}
	return app.meta
}

func (a *App) delPublish(appName string, streamName string) {
	pool := appName + "_" + streamName
	app, ok := a.List[pool]
	if !ok {
		return
	}
	for c := range app.client {
		close(c)
	}
	delete(a.List, pool)
}

func (a *App) delPlay(appName string, streamName string, client chan Pack) {
	pool := appName + "_" + streamName
	app, ok := a.List[pool]
	if !ok {
		return
	}
	delete(app.client, client)
	close(client)
}

func newApp() *App {
	app := &App{
		Gloab: make(map[string]listener),
		List:  make(map[string]*stream),
	}
	app.addGloab("flv", addFlvListen)
	return app
}

func addFlvListen(pools string) chan Pack {
	s := fmt.Sprint(time.Now().Unix())
	var flv av.FLV
	flv.GenFlv(pools + s)
	client := make(chan Pack)
	go func() {
		log.Println("addFlvListen - - - - > start")
		for pk := range client {
			flv.AddTag(pk.MessageTypeID, pk.Timestamp, pk.PayLoad)
		}
		defer flv.Close()
		log.Println("addFlvListen - - - - > end")
	}()
	return client
}
