package loges

import (
	"bytes"
	"encoding/xml"
	"net/url"
	"strconv"
	"strings"

	u "github.com/araddon/gou"
)

/*
	A monit listening service (http) that exracts monit info and forwards on

	fyi:  https://github.com/jeffcjohnson/monit_to_graphite
*/

// accepts a smokr message of monit format, and parses and converts to a name/value pair format
func MonitParse(data []byte) (url.Values, int64) {
	if idx := bytes.Index(data, []byte{'?'}); idx > 0 {
		data = data[idx+2:]
		r := MonitRequest{}
		//debug(string(data))
		err := xml.Unmarshal(data, &r)
		if err != nil {
			u.Log(u.ERROR, err)
			return nil, 0
		}
		return FlattenMonit(&r)
	}
	return nil, 0
}

// takes a monit object struture, and converts to a map[string]interface{}
// which it serializes to json, to send over a wire
func FlattenMonit(r *MonitRequest) (url.Values, int64) {
	m := make(url.Values)
	m.Set("host", r.Server.Host)
	host := "_" + r.Server.Host
	var ts int64
	for _, s := range r.Services {
		name := strings.Replace(s.Name, " ", "", -1)
		name = strings.Replace(name, host, "", -1)
		if s.Type == 3 {
			m.Set(name+".cpu.avg", s.Cpu.PercentTotalStr())
			m.Set(name+".mem.avg", strconv.Itoa(s.Memory.KilobyteTotal/1024))
			m.Set(name+".mempct.avg", s.Memory.PercentTotalStr())
		} else if s.Type == 0 {
			m.Set(name+".du.avg", s.Block.PercentTotalStr())
		} else if s.Type == 5 {
			// core os/cpu stuff
			m.Set(name+".cpu.avg", s.Cpu.PercentTotalStr())
			m.Set(name+".mem.avg", strconv.Itoa(s.Memory.KilobyteTotal/1024))
			m.Set(name+".mempct.avg", s.Memory.PercentTotalStr())
		}
		if ts == 0 {
			ts = int64(s.Ts) * 1e9
		}
	}
	return m, ts
}

type MonitServer struct {
	Uptime      int    `xml:"uptime"`
	Poll        int    `xml:"poll"`
	StartDelay  int    `xml:"startdelay"`
	Host        string `xml:"localhostname"`
	ControlFile string `xml:"controlfile"`
	//Httpd         map[string]int
	//Credentials   map[string]string
}
type MonitPlatform struct {
	Name    string `xml:"name"`
	Release string `xml:"release"`
	Version string `xml:"version"`
	Machine string `xml:"machine"`
	Cpu     int    `xml:"cpu"`
	Memory  int    `xml:"memory"`
	Swap    int    `xml:"swap"`
}
type MonitMemory struct {
	Percent       float32 `xml:"percent"`
	PercentTotal  float32 `xml:"percenttotal"`
	Kilobyte      int     `xml:"kilobyte"`
	KilobyteTotal int     `xml:"kilobytetotal"`
}

func (m *MonitMemory) PercentTotalStr() string {
	//strconv.FormatFloat(service.System.Load.Avg15, 'g', -1, 64), service.Collected_Sec}
	return strconv.FormatFloat(float64(m.PercentTotal), 'g', -1, 64)
}

type MonitCpu struct {
	Percent      float32 `xml:"percent"`
	PercentTotal float32 `xml:"percenttotal"`
}

func (m *MonitCpu) PercentTotalStr() string {
	//strconv.FormatFloat(service.System.Load.Avg15, 'g', -1, 64), service.Collected_Sec}
	return strconv.FormatFloat(float64(m.PercentTotal), 'g', -1, 64)
}

type MonitService struct {
	Type   int         `xml:"type"` // 0 = filesystem, 3 = service, 5= System(os)
	Name   string      `xml:"name,attr"`
	Ts     int         `xml:"collected_sec"`
	Tsu    int         `xml:"collected_usec"`
	Status int         `xml:"status"`
	Memory MonitMemory `xml:"memory"`
	Cpu    MonitCpu    `xml:"cpu"`
	Block  MonitBlock  `xml:"block"`
	INode  MonitBlock  `xml:"inode"`
	/*Status_hint   int
	Monitor       int
	MonitorMode   int
	PendingAction int
	Pid           int
	Ppid          int
	Uptime        int
	Children      int*/
}
type MonitRequest struct {
	Id            int            `xml:"id,attr"`
	Incarnation   string         `xml:"incarnation,attr"`
	Version       string         `xml:"version,attr"`
	Server        MonitServer    `xml:"server"`
	Platform      MonitPlatform  `xml:"platform"`
	Services      []MonitService `xml:"services>service"`
	ServiceGroups interface{}    `xml:"servicegroups"`
}

//  <block>
//     <percent>57.5</percent>
//     <usage>36786.6</usage>
//     <total>70127.3</total>
//  </block>
type MonitBlock struct {
	Percent float32 `xml:"percent"`
	Usage   float32 `xml:"usage"`
	Total   float32 `xml:"total"`
}

func (m *MonitBlock) PercentTotalStr() string {
	//strconv.FormatFloat(service.System.Load.Avg15, 'g', -1, 64), service.Collected_Sec}
	return strconv.FormatFloat(float64(m.Percent), 'g', -1, 64)
}
