package ping

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/go-ping/ping"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

//To add your node to the list simply add it to the nodes variable with the correct syntax and change the index number to relect the current number of nodes
//Your node can have a combined max fee of currently 2%, it will not be selected if it doesn't meet that requirement
//Your node needs to be pingable (Allow ICMP pings)
//Your node should be live 24/7 to avoid disrupting miners.

var nodes = []string{
	"fr1.vtconline.org",
	"p2proxy.vertcoin.org",
	"mindcraftblocks.com",
	"vtc-fl.javerity.com",
	"vtc-ca.javerity.com",
	"vtc.p2pminers.nl",
	"p2p-usa.xyz",
	"p2p-ekb.xyz",
	"asia.p2p-spb.xyz",
	"p2p-spb.xyz",
	"p2p-south.xyz",
	"siberia.mine.nu",
}
var NodeResults = make(map[time.Duration]string)

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
	MaxNextPing: 30 * time.Millisecond,
	PingPackets: 2,
	PingTimeout: 2000000000,
}

type SelectedNode struct {
	P2PoolStratum string
	P2PoolURL     string
}

var Node SelectedNode

func GetNode(testnet bool) {
	if testnet {
		Node = SelectedNode{
			P2PoolStratum: networks.Active.P2ProxyStratum,
			P2PoolURL:     networks.Active.P2ProxyURL,
		}
	} else {
		Node = SelectedNode{
			P2PoolStratum: getClosestNodeStratum(),
			P2PoolURL:     getClosestNodeURL(),
		}
	}
}

var selectedNode string = selector()

func selector() (selected string) {
	_, err := getNodeInformation("127.0.0.1:9171")
	if err == nil {
		logging.Infof("Selected local p2pool node\n")
		selected = "127.0.0.1:9171"
		return selected
	} else {
		logging.Infof("No local node detected, selecting other public nodes\n")
	}

	err = pingNodes()
	if err != nil {
		logging.Warnf("Nodes could not be pinged, selecting random node")
		rand.Seed(time.Now().Unix())
		selected = nodes[rand.Intn(len(nodes))]
		logging.Infof("%s has been randomly selected", selected)
		return selected
	}
	SortedPingSlice := SortPingtimes()

	for i := 0; i < len(SortedPingSlice); i++ {
		NodeInformation, _ := getNodeInformation(NodeResults[SortedPingSlice[i]])
		Fee := CheckFee(NodeInformation)
		if Fee {
			CurrentMiners := CheckCurrentMiners(NodeInformation)
			if CurrentMiners {
				selected = NodeResults[SortedPingSlice[i]]
				logging.Infof("%s selected, fulfilled all requirements\n", NodeResults[SortedPingSlice[i]])
				break
			} else { //If the node fulfills the max fee requirement but there is more than the set MaxNextping to the next node it will select the current
				DetermineNextPingTime := SortedPingSlice[i+1]
				DetermineNextPingTime -= SortedPingSlice[i]
				if DetermineNextPingTime > Set.MaxNextPing {
					selected = NodeResults[SortedPingSlice[i]]
					logging.Infof("%s selected, next node has too high ping time\n", NodeResults[SortedPingSlice[i]])
					break
				}
				logging.Warnf("%s had more than %v miners, trying new inorder to retain efficiency\n", SortedPingSlice[i], Set.MaxMiners)
			}
		} else {
			logging.Warnf("%s had more than a %f fee, trying new\n", SortedPingSlice[i], Set.MaxFee)
		}
	}
	return selected
}

func pingNodes() error {
	results := make([]time.Duration, len(nodes))

	for i := 0; i < len(nodes); i++ {
		pinger, err := ping.NewPinger(nodes[i])
		pinger.SetPrivileged(true)       //This line is needed for windows because of ICMP
		pinger.Timeout = Set.PingTimeout //Sets the time for which the pinger will timeout regardless of how many packets there has been recieved
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
			return err
		}
		pinger.Count = Set.PingPackets //Number of packets to be sent to each node
		err = pinger.Run()
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
			return err
		}
		if nodes[i] == "p2proxy.vertcoin.org" { //Currently this node uses port 9172 instead of 9171, if this changes this statement can be removed and 9171 can be added to all nodes
			nodes[i] += ":9172"
		} else {
			nodes[i] += ":9171"
		}
		results[i] = pinger.Statistics().AvgRtt
		NodeResults[results[i]] = nodes[i]
		logging.Infof("%s: %v \n", nodes[i], results[i])
	}
	return nil
}

func SortPingtimes() []time.Duration {
	SortedMap := make([]time.Duration, 0, len(NodeResults))
	for k := range NodeResults {
		SortedMap = append(SortedMap, k)
	}
	sort.Slice(SortedMap, func(i, j int) bool { return SortedMap[i] < SortedMap[j] })
	return SortedMap
}

func getNodeInformation(Node string) (jsonPayload map[string]interface{}, err error) {
	NodeURL := "http://"
	NodeURL += Node
	err = util.GetJson(fmt.Sprintf("%s/local_stats", NodeURL), &jsonPayload)
	if err != nil {
		logging.Errorf("Unable to fetch node information\n", err.Error())
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
	if fee > Set.MaxFee {
		return false
	}
	return true
}

//To ensure efficiency of the selected p2pool node a limit of miners has been put in place, returns true if the number is equal to Maxminers or below
func CheckCurrentMiners(jsonPayload map[string]interface{}) bool {
	CurrentMiners, ok := jsonPayload["miner_hash_rates"].(string)
	if !ok {
		return false
	}
	if len(CurrentMiners) > Set.MaxMiners {
		return false
	}
	return true
}

func getClosestNodeStratum() (stratum string) {
	stratum = "stratum+tcp://"
	stratum += selectedNode
	return stratum
}

func getClosestNodeURL() (URL string) {
	URL = "http://"
	URL += selectedNode
	URL += "/"
	return URL
}
