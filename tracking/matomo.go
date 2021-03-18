package tracking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/vertiond/verthash-one-click-miner/logging"
	"github.com/vertiond/verthash-one-click-miner/util"
)

type TrackingRequest struct {
	Category string
	Action   string
	Name     string
}

var trackChan chan TrackingRequest
var trackingEnabled bool
var waitGroup sync.WaitGroup

func Enable() {
	trackingEnabled = true
	saveState()
}

func IsEnabled() bool {
	return trackingEnabled
}

func Disable() {
	trackingEnabled = false
	saveState()
}

func loadState() {
	dat, err := ioutil.ReadFile(filepath.Join(util.DataDirectory(), "tracking"))
	if err != nil {
		// // Default enable tracking
		// Enable()
		// Default disable tracking
		Disable()
		return
	}
	trackingEnabled = (string(dat) == "1")
	if trackingEnabled {
		Disable()
	}
}

func saveState() {
	value := "1"
	if !trackingEnabled {
		value = "0"
	}
	ioutil.WriteFile(filepath.Join(util.DataDirectory(), "tracking"), []byte(value), 0644)
}

func StartTracker() {
	matomoClient := &http.Client{Timeout: 2 * time.Second}
	trackChan = make(chan TrackingRequest, 100)
	waitGroup = sync.WaitGroup{}
	loadState()
	new := true
	go func() {
		waitGroup.Add(1)
		for t := range trackChan {
			if !trackingEnabled {
				continue
			}
			req, err := http.NewRequest("GET", "https://matomo.gertjaap.org/matomo.php", nil)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}

			q := req.URL.Query()
			q.Add("idsite", "2")
			q.Add("rec", "1")
			if new {
				new = false
				q.Add("new_visit", "1")
			}
			q.Add("action_name", fmt.Sprintf("%s/%s", t.Category, t.Action))
			q.Add("e_c", t.Category)
			q.Add("e_a", t.Action)
			q.Add("e_n", t.Name)
			q.Add("ua", userAgent())
			q.Add("_id", visitorId())
			req.URL.RawQuery = q.Encode()

			r, err := matomoClient.Do(req)
			if err != nil {
				logging.Warnf("Error sending tracking data: %v", err)
				continue
			}
			if r.Body != nil {
				r.Body.Close()
			}
		}
		waitGroup.Done()
	}()
}

func Stop() {
	close(trackChan)
	waitGroup.Wait()
}

func Track(req TrackingRequest) {
	trackChan <- req
}

var vId string

func visitorId() string {
	if vId == "" {
		dat, err := ioutil.ReadFile(filepath.Join(util.DataDirectory(), "unique_id"))
		if err != nil {
			dat = make([]byte, 8)
			rand.Read(dat)
			ioutil.WriteFile(filepath.Join(util.DataDirectory(), "unique_id"), dat, 0644)
		}
		vId = strings.ToUpper(hex.EncodeToString(dat))
	}
	return vId
}

var ua string

func userAgent() string {
	if ua == "" {
		ua = fmt.Sprintf("OCM/%s %s/%s", GetVersion(), runtime.GOOS, runtime.GOARCH)
	}
	return ua
}

func GetVersion() string {
	return version
}
