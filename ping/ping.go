package ping

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/go-ping/ping"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
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
}

var Selected SelectedNode

func GetSelectedNode(testnet bool) {
	if testnet {
		Selected = SelectedNode{
			P2PoolStratum: networks.Active.P2ProxyStratum,
			P2PoolURL:     networks.Active.P2ProxyURL,
		}
	} else {
		selector()
	}
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
		err = PingNodes(NodeList)
		if err != nil {
			logging.Warnf("Nodes could not be pinged, selecting random node\n")
		} else {

			sort.Slice(NodeList, func(i, j int) bool {
				return NodeList[i].PingTime < NodeList[j].PingTime
			})

			for i := 0; i < len(NodeList); i++ {
				if NodeList[i].PingTime == 0 { // We need to skip nodes with a pingTime of 0ms, they're either not active or they're not responding to pings and OCM will select it if it responds to the following requests.
					continue
				}
				nodeInformation, _ := GetNodeInformation(NodeList[i].URL)
				fee := CheckFee(nodeInformation)
				if fee {
					currentMiners := CheckCurrentMiners(nodeInformation)
					if currentMiners {
						Selected = SelectedNode{
							P2PoolStratum: NodeList[i].Stratum,
							P2PoolURL:     NodeList[i].URL,
						}
						logging.Infof("%s selected, fulfilled all requirements\n", NodeList[i].Hostname)
						break
					} else { //If the node fulfills the max fee requirement but there is more than the set MaxNextping to the next node it will select the current
						// The last node has no next node to compare against, so take it.
						if i+1 >= len(NodeList) || NodeList[i+1].PingTime-NodeList[i].PingTime > Set.MaxNextPing {
							Selected = SelectedNode{
								P2PoolStratum: NodeList[i].Stratum,
								P2PoolURL:     NodeList[i].URL,
							}
							logging.Infof("%s selected, next node has too high ping time\n", NodeList[i].Hostname)
							break
						}
						logging.Warnf("%s had more than %v miners, trying new inorder to retain efficiency\n", NodeList[i].Hostname, Set.MaxMiners)
					}
				} else {
					logging.Warnf("%s is either unreachable or had more than a %f fee, trying new\n", NodeList[i].Hostname, Set.MaxFee)
				}
			}
		}

		// Nothing was picked by ping time, so take a random node that is up.
		if Selected.P2PoolURL == "" {
			if node, ok := pickResponsiveNode(NodeList, nodeResponds); ok {
				Selected = SelectedNode{
					P2PoolStratum: node.Stratum,
					P2PoolURL:     node.URL,
				}
				logging.Infof("%s has been randomly selected\n", node.Hostname)
			} else {
				logging.Warnf("No p2pool node responded, using fallback node\n")
				useFallbackNode()
			}
		}
	}
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

func nodeResponds(url string) bool {
	_, err := GetNodeInformation(url)
	return err == nil
}

// Instead of making a http request to the node each time we need to get information, we do it once and then reuse the collected data.
func GetNodeInformation(NodeURL string) (jsonPayload map[string]interface{}, err error) {
	err = util.GetJson(fmt.Sprintf("%slocal_stats", NodeURL), &jsonPayload)
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
	currentMiners, _ := jsonPayload["miner_hash_rates"].(string)
	return len(currentMiners) <= Set.MaxMiners
}
