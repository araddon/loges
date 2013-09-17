package loges

import (
	"bytes"
	"fmt"
	u "github.com/araddon/gou"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NvMetrics struct {
	url.Values
}

func NewNvMetrics(qs string) (NvMetrics, error) {
	if li := strings.LastIndex(qs, ","); li > 0 {
		vals := strings.Split(qs, ",")
		nv := make(url.Values)
		for _, val := range vals {
			if parts := strings.Split(strings.Trim(val, " \n\r"), " "); len(parts) == 2 {
				nv.Set(parts[0], parts[1])
			}
		}
		//u.Debug(qs, " ---- ", nv)
		return NvMetrics{nv}, nil
	}
	nv, err := url.ParseQuery(qs)
	return NvMetrics{nv}, err
}

func (uv *NvMetrics) MetricTypeVal(name string) (string, string) {
	if v := uv.Values.Get(name); len(v) > 0 {
		//u.Debug(name, "---", v)
		if li := strings.LastIndex(name, "."); li > 0 {
			return name[li+1:], v
		}
	}
	return "", ""
}

func GraphiteTransform(addr, prefix string) LineTransform {
	ticker := time.NewTicker(time.Second * 60)
	var mu sync.Mutex
	buf := &bytes.Buffer{}
	go func() {
		for {
			select {
			case <-ticker.C:
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					u.Errorf("Failed to connect to graphite/carbon: %+v", err)
				} else {
					//u.Infof("Connected graphite to %v", addr)
					mu.Lock()
					//u.Debug(string(buf.Bytes()))
					io.Copy(conn, buf)
					mu.Unlock()
				}
				if conn != nil {
					conn.Close()
				}
				//case <-stopper:
				//	return
			}

		}
	}()

	return func(d *LineEvent) *Event {
		//u.Debugf("ll=%s   %s", d.DataType, string(d.Data))
		if d.DataType == "METRIC" || d.DataType == "METR" {
			line := string(d.Data)
			tsStr := strconv.FormatInt(time.Now().In(time.UTC).Unix(), 10)
			if iMetric := strings.Index(line, d.DataType); iMetric > 0 {
				line = line[iMetric+len(d.DataType)+1:]
				line = strings.Trim(line, " ")
			}
			// 1.  Read nv/pairs
			nv, err := NewNvMetrics(line)
			if err != nil {
				u.Error(err)
				return nil
			}
			host := nv.Get("host")
			if len(host) == 0 {
				host = hostName
			}
			//u.Debugf("To Graphite! h='%s'  data=%s", host, string(d.Data))
			mu.Lock()
			defer mu.Unlock()
			// 2.  parse the .avg, .ct and switch
			for n, _ := range nv.Values {
				switch metType, val := nv.MetricTypeVal(n); metType {
				case "avg": // Gauge
					//n = strings.Replace(n, ".avg", "", -1)
					if _, err = fmt.Fprintf(buf, "%s.%s.%s %s %s\n", prefix, host, n, val, tsStr); err != nil {
						u.Error(err)
						return nil
					}
				case "ct":
					n = strings.Replace(n, ".ct", ".count", -1)
					if _, err = fmt.Fprintf(buf, "%s.%s.%s %s %s\n", prefix, host, n, val, tsStr); err != nil {
						u.Error(err)
						return nil
					}
				default:
					// ?
				}
			}
		}
		return nil
	}
}

type GraphiteRunner struct {
	addr string
	conn net.Conn
	Mu   sync.Mutex
}

func NewGraphiteRunner(addr string) *GraphiteRunner {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		u.Errorf("Failed to connect to graphite/carbon: %+v", err)
	} else {
		u.Infof("Connected graphite to %v", addr)
	}
	return &GraphiteRunner{conn: conn}
}

func (g *GraphiteRunner) Close() {
	if g.conn != nil {
		g.conn.Close()
	}
}
