package loges

import (
	"bytes"
	"errors"
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
func (uv *NvMetrics) Value(name string) (interface{}, error) {
	if v := uv.Values.Get(name); len(v) > 0 {
		//u.Debug(name, "---", v)
		if li := strings.LastIndex(name, "."); li > 0 {
			metType := name[li+1:]
			switch metType {
			case "avg", "pct": // Gauge, Timer
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					return f, nil
				} else {
					u.Error(err)
					return nil, err
				}
			case "ct":
				if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
					return iv, nil
				} else {
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						return int64(f), nil
					} else {
						u.Errorf(`Could not parse integer or float for   "%v.ct" v=%v`, name, v)
						return nil, err
					}
				}
			case "value":
				if fv, err := strconv.ParseFloat(v, 64); err == nil {
					return int64(fv), nil
				} else {
					if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
						return iv, nil
					} else {
						u.Errorf(`Could not parse integer or float for   "%v.ct" v=%v`, name, v)
						return nil, err
					}
				}
			}
		}
	}
	return nil, errors.New("not converted")
}
func GraphiteTransform(logstashType, addr, prefix string, metricsToEs bool) LineTransform {
	ticker := time.NewTicker(time.Second * 60)
	loc := time.UTC
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
		if d.IsMetric() {
			line := string(d.Data)
			line = strings.Trim(line, " ")
			tsStr := strconv.FormatInt(time.Now().In(loc).Unix(), 10)

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

			evt := NewTsEvent(logstashType, d.Source, line, time.Now().In(loc))
			evt.Fields = make(map[string]interface{})
			evt.Fields["host"] = hostName
			evt.Fields["level"] = d.LogLevel
			//u.Debugf("To Graphite! h='%s'  data=%s", host, string(d.Data))
			mu.Lock()
			defer mu.Unlock()
			// 2.  parse the .avg, .ct and switch
			for n, _ := range nv.Values {
				metType, val := nv.MetricTypeVal(n)
				if metVal, err := nv.Value(n); err == nil {
					grapiteName := strings.Replace(n, ".", "_", -1)
					evt.Fields[grapiteName] = metVal
				} else {
					continue
				}
				switch metType {
				case "avg", "pct": // Gauge, Timer
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
				case "value":
					n = strings.Replace(n, ".value", ".last", -1)
					if _, err = fmt.Fprintf(buf, "%s.%s.%s %s %s\n", prefix, host, n, val, tsStr); err != nil {
						u.Warnf("Could not convert value:  %v:%v", n, val)
						//return nil
					}
				default:
					// ?
					u.Warnf("could not recognize: %v", line)
				}
			}
			if metricsToEs {
				return evt
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
