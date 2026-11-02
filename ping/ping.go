package ping

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-ping/ping"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

type Conditions struct {
	MaxFee      float64
	MaxMiners   int
	MaxNextPing time.Duration
	PingPackets int
	PingTimeout time.Duration
}

var Set = Conditions{
	MaxFee:      2,
	MaxMiners:   35,
	MaxNextPing: 25 * time.Millisecond, //Prevents the miner from using a node with too high latency compared to other nodes. Should prevent the miner from using nodes outside of their region
	PingPackets: 2,
	PingTimeout: 2000000000,
}

type SelectedNode struct {
	P2PoolStratum string
	P2PoolURL     string
}

type Nodes struct {
	Hostname string `json:"Hostname"`
	Stratum  string `json:"Stratum"`
	URL      string `json:"URL"`
	PingTime time.Duration
	info     map[string]interface{} // local_stats, nil if the node didn't answer
}

var Selected SelectedNode

func GetSelectedNode() {
	selector()
}

func selector() {
	_, err := GetNodeInformation("http://127.0.0.1:9171/")
	if err == nil {
		logging.Infof("Selected local p2pool node\n")
		Selected = SelectedNode{
			P2PoolStratum: "stratum+tcp://127.0.0.1:9171",
			P2PoolURL:     "http://127.0.0.1:9171/",
		}
	} else {
		logging.Infof("No local node detected, selecting other public nodes\n")

		NodeList := []Nodes{}
		err = util.GetJson("https://raw.githubusercontent.com/vertcoin-project/one-click-miner-vnext/master/p2pool_nodes.json", &NodeList)

		// If the node list can't be fetched, fall back to a known node instead of
		// picking from an empty list.
		if err != nil || len(NodeList) == 0 {
			logging.Warnf("P2pool nodes could not be fetched, using fallback node\n")
			useFallbackNode()
			return
		}

		Selected = SelectedNode{}

		// Fetch every node's stats while pinging, so a dead node costs at most
		// one timeout in total instead of one per node.
		var wg sync.WaitGroup
		for i := range NodeList {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if info, err := GetNodeInformation(NodeList[i].URL); err == nil {
					NodeList[i].info = info
				}
			}(i)
		}
		err = PingNodes(NodeList)
		wg.Wait()

		if sel, ok := pick(NodeList, err); ok {
			Selected = sel
		} else {
			useFallbackNode()
		}
	}
}

// pick chooses a node from fetched and pinged nodes: the lowest ping that
// passes the fee and miner checks, then a random responsive node. pingErr
// means no node could be pinged. It returns false when no node is usable.
func pick(NodeList []Nodes, pingErr error) (SelectedNode, bool) {
	selected := func(n Nodes) SelectedNode {
		return SelectedNode{P2PoolStratum: n.Stratum, P2PoolURL: n.URL}
	}
	usable := func(n Nodes) bool {
		return n.info != nil && CheckFee(n.info) && CheckCurrentMiners(n.info)
	}

	if pingErr != nil {
		logging.Warnf("Nodes could not be pinged, selecting random node\n")
	} else {
		// Lowest ping first. Nodes that block ping (PingTime 0) go last but
		// can still be picked if their stats answered.
		sort.SliceStable(NodeList, func(i, j int) bool {
			a, b := NodeList[i].PingTime, NodeList[j].PingTime
			if a == 0 || b == 0 {
				return b == 0 && a != 0
			}
			return a < b
		})

		for i := 0; i < len(NodeList); i++ {
			if NodeList[i].info == nil {
				continue
			}
			if !CheckFee(NodeList[i].info) {
				logging.Warnf("%s is either unreachable or had more than a %f fee, trying new\n", NodeList[i].Hostname, Set.MaxFee)
				continue
			}
			if CheckCurrentMiners(NodeList[i].info) {
				logging.Infof("%s selected, fulfilled all requirements\n", NodeList[i].Hostname)
				return selected(NodeList[i]), true
			}
			// Too many miners: move on only if a usable node is close by in
			// ping. Otherwise this node is still the best choice.
			alternative := -1
			for j := i + 1; j < len(NodeList); j++ {
				if NodeList[j].PingTime > 0 && usable(NodeList[j]) {
					alternative = j
					break
				}
			}
			if alternative < 0 || NodeList[i].PingTime == 0 || NodeList[alternative].PingTime-NodeList[i].PingTime > Set.MaxNextPing {
				logging.Infof("%s selected, no usable node close by in ping\n", NodeList[i].Hostname)
				return selected(NodeList[i]), true
			}
			logging.Warnf("%s had more than %v miners, trying new inorder to retain efficiency\n", NodeList[i].Hostname, Set.MaxMiners)
		}
	}

	// Nothing was picked by ping time, so take a random node that is up.
	up := map[string]bool{}
	for _, n := range NodeList {
		up[n.URL] = n.info != nil
	}
	if node, ok := pickResponsiveNode(NodeList, func(url string) bool { return up[url] }); ok {
		logging.Infof("%s has been randomly selected\n", node.Hostname)
		return selected(node), true
	}
	logging.Warnf("No p2pool node responded, using fallback node\n")
	return SelectedNode{}, false
}

func useFallbackNode() {
	Selected = SelectedNode{
		P2PoolStratum: "stratum+tcp://vtc-fl.javerity.com:9171",
		P2PoolURL:     "http://vtc-fl.javerity.com:9171/",
	}
}

// PingNodes sets PingTime on every node that answers. A node that can't be
// resolved or pinged keeps a PingTime of 0 and is skipped by the selector;
// an error is only returned when no node could be pinged at all.
func PingNodes(NodeList []Nodes) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var lastErr error
	pinged := 0
	for i := 0; i < len(NodeList); i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			pinger, err := ping.NewPinger(NodeList[i].Hostname)
			if err != nil {
				logging.Warnf("%s could not be pinged: %v\n", NodeList[i].Hostname, err)
				mu.Lock()
				lastErr = err
				mu.Unlock()
				return
			}
			pinger.SetPrivileged(true)       //This line is needed for windows because of ICMP
			pinger.Timeout = Set.PingTimeout //Sets the time for which the pinger will timeout regardless of how many packets there has been recieved
			pinger.Count = Set.PingPackets   //Number of packets to be sent to each node
			err = pinger.Run()
			if err != nil {
				logging.Warnf("%s could not be pinged: %v\n", NodeList[i].Hostname, err)
				mu.Lock()
				lastErr = err
				mu.Unlock()
				return
			}
			NodeList[i].PingTime = pinger.Statistics().AvgRtt
			logging.Infof("%s: %v \n", NodeList[i].Hostname, NodeList[i].PingTime)
			mu.Lock()
			pinged++
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	if pinged == 0 {
		if lastErr == nil {
			lastErr = fmt.Errorf("no node answered")
		}
		return lastErr
	}
	return nil
}

// pickResponsiveNode returns a random node whose stats page responds, or false
// if none do. Used when ping times can't help pick a node.
func pickResponsiveNode(NodeList []Nodes, responds func(url string) bool) (Nodes, bool) {
	for _, i := range rand.Perm(len(NodeList)) {
		if responds(NodeList[i].URL) {
			return NodeList[i], true
		}
		logging.Warnf("%s is unreachable, trying another node\n", NodeList[i].Hostname)
	}
	return Nodes{}, false
}

// A healthy node answers in well under a second, so a short timeout keeps a
// dead node from stalling node selection.
var nodeClient = &http.Client{Timeout: 3 * time.Second}

// Instead of making a http request to the node each time we need to get information, we do it once and then reuse the collected data.
func GetNodeInformation(NodeURL string) (jsonPayload map[string]interface{}, err error) {
	r, err := nodeClient.Get(fmt.Sprintf("%slocal_stats", NodeURL))
	if err == nil {
		defer r.Body.Close()
		err = json.NewDecoder(r.Body).Decode(&jsonPayload)
	}
	if err != nil {
		if NodeURL != "http://127.0.0.1:9171/" {
			logging.Errorf("Unable to fetch node information: %s\n", err.Error())
		}
		return jsonPayload, err
	}
	return jsonPayload, nil
}

func CheckFee(jsonPayload map[string]interface{}) bool {
	fee, ok := jsonPayload["fee"].(float64)
	if !ok {
		return false
	}
	donationFee, ok := jsonPayload["donation_proportion"].(float64)
	if !ok {
		return false
	}
	fee += donationFee
	return fee <= Set.MaxFee
}

// To ensure efficiency of the selected p2pool node a limit of miners has been put in place, returns true if the number is equal to Maxminers or below
func CheckCurrentMiners(jsonPayload map[string]interface{}) bool {
	// miner_hash_rates is an object of payout address to hashrate, one entry per miner.
	currentMiners, _ := jsonPayload["miner_hash_rates"].(map[string]interface{})
	return len(currentMiners) <= Set.MaxMiners
}
