package tracking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
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
	dat, err := os.ReadFile(filepath.Join(util.DataDirectory(), "tracking"))
	if err != nil {
		Enable()
		return
	}
	trackingEnabled = (string(dat) == "1")
}

func saveState() {
	value := "1"
	if !trackingEnabled {
		value = "0"
	}
	err := os.WriteFile(filepath.Join(util.DataDirectory(), "tracking"), []byte(value), 0644)
	if err != nil {
		logging.Errorf("Error writing tracking state: %v", err)
	}
}

func StartTracker() {
	matomoClient := &http.Client{Timeout: 2 * time.Second}
	trackChan = make(chan TrackingRequest, 100)
	waitGroup = sync.WaitGroup{}
	loadState()
	new := true
	waitGroup.Add(1)
	go func() {
		for t := range trackChan {
			if !trackingEnabled {
				continue
			}
			req, err := http.NewRequest("GET", "https://analytics.javerity.com/matomo.php", nil)
			if err != nil {
				log.Print(err)
				os.Exit(1)
			}

			q := req.URL.Query()
			q.Add("idsite", "3")
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
		dat, err := os.ReadFile(filepath.Join(util.DataDirectory(), "unique_id"))
		if err != nil {
			dat = make([]byte, 8)
			_, err := rand.Read(dat)
			if err != nil {
				logging.Errorf("Error reading random ID: %v", err)
			}
			err = os.WriteFile(filepath.Join(util.DataDirectory(), "unique_id"), dat, 0644)
			if err != nil {
				logging.Errorf("Error writing random ID: %v", err)
			}

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
