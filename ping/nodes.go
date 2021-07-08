package ping

import (
	"time"
)

type Node struct {
	Hostname string
	Stratum  string
	URL      string
	PingTime time.Duration
}

//To add your node to the list simply add it to the slice below with following the same syntax.
//Your node can have a combined max fee of 2%, it will not be selected if it doesn't meet that requirement.
//Your node needs to be pingable (Allow ICMP pings).
//Your node should be live 24/7 to avoid disrupting miners.

var NodeList = []Node{
	{
		Hostname: "p2proxy.vertcoin.org",
		Stratum:  "stratum+tcp://p2proxy.vertcoin.org:9172",
		URL:      "http://p2proxy.vertcoin.org:9172/",
	},
	{
		Hostname: "mindcraftblocks.com",
		Stratum:  "stratum+tcp://mindcraftblocks.com:9171",
		URL:      "http://mindcraftblocks.com:9171/",
	},
	{
		Hostname: "vtc-fl.javerity.com",
		Stratum:  "stratum+tcp://vtc-fl.javerity.com:9171",
		URL:      "http://vtc-fl.javerity.com:9171/",
	},
	{
		Hostname: "vtc-ca.javerity.com",
		Stratum:  "stratum+tcp://vtc-ca.javerity.com:9171",
		URL:      "http://vtc-ca.javerity.com:9171/",
	},
	{
		Hostname: "vtc.p2pminers.nl",
		Stratum:  "stratum+tcp://vtc.p2pminers.nl:9171",
		URL:      "http://vtc.p2pminers.nl:9171/",
	},
	{
		Hostname: "p2p-usa.xyz",
		Stratum:  "stratum+tcp://p2p-usa.xyz:9171",
		URL:      "http://p2p-usa.xyz:9171/",
	},
	{
		Hostname: "p2p-ekb.xyz",
		Stratum:  "stratum+tcp://p2p-ekb.xyz:9171",
		URL:      "http://p2p-ekb.xyz:9171/",
	},
	{
		Hostname: "asia.p2p-spb.xyz",
		Stratum:  "stratum+tcp://asia.p2p-spb.xyz:9171",
		URL:      "http://asia.p2p-spb.xyz:9171/",
	},
	{
		Hostname: "p2p-spb.xyz",
		Stratum:  "stratum+tcp://p2p-spb.xyz:9171",
		URL:      "http://p2p-spb.xyz:9171/",
	},
	{
		Hostname: "p2p-south.xyz",
		Stratum:  "stratum+tcp://p2p-south.xyz:9171",
		URL:      "http://p2p-south.xyz:9171/",
	},
	{
		Hostname: "siberia.mine.nu",
		Stratum:  "stratum+tcp://siberia.mine.nu:9171",
		URL:      "http://siberia.mine.nu:9171/",
	},
}
