package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/twsnmp/twlogeye/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type twLogEyeNotifyEnt struct {
	Time  string
	ID    string
	Level string
	Title string
	Tags  string
	Src   string
	Log   string
}

type twLogEyeLogEnt struct {
	Time string
	Src  string
	Log  string
}

type twLogEyeSyslogReport struct {
	Time        string
	Normal      int32
	Warn        int32
	Error       int32
	Patterns    int32
	ErrPatterns int32
}

type twLogEyeTrapReport struct {
	Time  string
	Count int32
	Types int32
}

type twLogEyeNetflowReport struct {
	Time      string
	Packets   int64
	Bytes     int64
	MACs      int32
	IPs       int32
	Flows     int32
	Protocols int32
	Fumbles   int32
	Hosts     int32
	Locs      int32
	Country   int32
}

type twLogEyeWindowsEventReport struct {
	Time       string
	Normal     int32
	Warn       int32
	Error      int32
	Types      int32
	ErrorTypes int32
}

type twLogEyeOTelReport struct {
	Time        string
	Normal      int32
	Warn        int32
	Error       int32
	Types       int32
	ErrorTypes  int32
	Hosts       int32
	TraceIds    int32
	TraceCount  int32
	MericsCount int32
}

type twLogEyeMqttReport struct {
	Time  string
	Count int32
	Types int32
}

type twLogEyeMonitorReport struct {
	Time    string
	CPU     float64
	Memory  float64
	Load    float64
	Disk    float64
	Net     float64
	Bytes   int64
	DBSpeed float64
	DBSize  int64
}

type twLogEyeAnomalyReport struct {
	Time  string
	Type  string
	Score float64
}

func (b *App) readLogFromTwLogEye(lf *LogFile) error {
	src := lf.LogSrc
	target := src.Target
	if target == "" {
		target = "notify"
	}
	subTarget := src.SubTarget
	anomalyType := src.ReportType
	if anomalyType == "" {
		anomalyType = "monitor"
	}

	if u, err := url.Parse(src.Server); err == nil && u.Path != "" {
		p := strings.Trim(u.Path, "/")
		parts := strings.Split(p, "/")
		if len(parts) > 0 && parts[0] != "" {
			target = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			subTarget = parts[1]
		}
		if len(parts) > 2 && parts[2] != "" {
			anomalyType = parts[2]
		}
	}

	client, conn, err := b.getTwLogEyeClient(src)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
	}

	st, et := b.getLogSourceTimeRange(src)
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		switch target {
		case "notify":
			s, err := client.SearchNotify(context.Background(), &api.NofifyRequest{
				Start: st,
				End:   et,
				Level: src.Level,
			})
			if err != nil {
				OutLog("SearchNotify err=%v", err)
				return
			}
			for {
				if b.stopProcess {
					return
				}
				r, err := s.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					OutLog("SearchNotify recv err=%v", err)
					return
				}
				j, err := json.Marshal(&twLogEyeNotifyEnt{
					Time:  getTimeStr(r.GetTime()),
					Src:   r.GetSrc(),
					Level: r.GetLevel(),
					ID:    r.GetId(),
					Tags:  r.GetTags(),
					Title: r.GetTitle(),
					Log:   r.GetLog(),
				})
				if err != nil {
					continue
				}
				if _, err := pw.Write(append(j, '\n')); err != nil {
					return
				}
			}

		case "logs":
			s, err := client.SearchLog(context.Background(), &api.LogRequest{
				Logtype: subTarget,
				Start:   st,
				End:     et,
				Search:  src.Pattern,
			})
			if err != nil {
				OutLog("SearchLog err=%v", err)
				return
			}
			for {
				if b.stopProcess {
					return
				}
				r, err := s.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					OutLog("SearchLog recv err=%v", err)
					return
				}
				j, err := json.Marshal(&twLogEyeLogEnt{
					Time: getTimeStr(r.GetTime()),
					Src:  r.GetSrc(),
					Log:  r.GetLog(),
				})
				if err != nil {
					continue
				}
				if _, err := pw.Write(append(j, '\n')); err != nil {
					return
				}
			}

		case "report":
			b.streamTwLogEyeReport(client, pw, subTarget, anomalyType, st, et)
		}
	}()

	b.readOneLogFile(lf, pr)
	return nil
}

func (b *App) streamTwLogEyeReport(client api.TWLogEyeServiceClient, pw *io.PipeWriter, subTarget, anomalyType string, st, et int64) {
	ctx := context.Background()
	switch subTarget {
	case "syslog":
		s, err := client.GetSyslogReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeSyslogReport{
				Time:        getTimeStr(r.GetTime()),
				Normal:      r.GetNormal(),
				Warn:        r.GetWarn(),
				Error:       r.GetError(),
				Patterns:    r.GetPatterns(),
				ErrPatterns: r.GetErrPatterns(),
			})
			pw.Write(append(j, '\n'))
		}
	case "trap":
		s, err := client.GetTrapReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeTrapReport{
				Time:  getTimeStr(r.GetTime()),
				Count: r.GetCount(),
				Types: r.GetTypes(),
			})
			pw.Write(append(j, '\n'))
		}
	case "netflow":
		s, err := client.GetNetflowReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeNetflowReport{
				Time:      getTimeStr(r.GetTime()),
				Packets:   r.GetPackets(),
				Bytes:     r.GetBytes(),
				MACs:      r.GetMacs(),
				IPs:       r.GetIps(),
				Flows:     r.GetFlows(),
				Protocols: r.GetProtocols(),
				Fumbles:   r.GetFumbles(),
				Hosts:     r.GetHosts(),
				Locs:      r.GetLocs(),
				Country:   r.GetCountry(),
			})
			pw.Write(append(j, '\n'))
		}
	case "winevent":
		s, err := client.GetWindowsEventReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeWindowsEventReport{
				Time:       getTimeStr(r.GetTime()),
				Normal:     r.GetNormal(),
				Warn:       r.GetWarn(),
				Error:      r.GetError(),
				Types:      r.GetTypes(),
				ErrorTypes: r.GetErrorTypes(),
			})
			pw.Write(append(j, '\n'))
		}
	case "otel":
		s, err := client.GetOTelReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeOTelReport{
				Time:        getTimeStr(r.GetTime()),
				Normal:      r.GetNormal(),
				Warn:        r.GetWarn(),
				Error:       r.GetError(),
				Types:       r.GetTypes(),
				ErrorTypes:  r.GetErrorTypes(),
				Hosts:       r.GetHosts(),
				TraceIds:    r.GetTraceIds(),
				TraceCount:  r.GetTraceCount(),
				MericsCount: r.GetMericsCount(),
			})
			pw.Write(append(j, '\n'))
		}
	case "mqtt":
		s, err := client.GetMqttReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeMqttReport{
				Time:  getTimeStr(r.GetTime()),
				Count: r.GetCount(),
				Types: r.GetTypes(),
			})
			pw.Write(append(j, '\n'))
		}
	case "monitor":
		s, err := client.GetMonitorReport(ctx, &api.ReportRequest{Start: st, End: et})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeMonitorReport{
				Time:    getTimeStr(r.GetTime()),
				CPU:     r.GetCpu(),
				Memory:  r.GetMemory(),
				Load:    r.GetLoad(),
				Disk:    r.GetDisk(),
				Net:     r.GetNet(),
				Bytes:   r.GetBytes(),
				DBSpeed: r.GetDbSpeed(),
				DBSize:  r.GetDbSize(),
			})
			pw.Write(append(j, '\n'))
		}
	case "anomaly":
		s, err := client.GetAnomalyReport(ctx, &api.AnomalyReportRequest{Start: st, End: et, Type: anomalyType})
		if err != nil {
			return
		}
		for {
			if b.stopProcess {
				return
			}
			r, err := s.Recv()
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			j, _ := json.Marshal(&twLogEyeAnomalyReport{
				Time:  getTimeStr(r.GetTime()),
				Type:  anomalyType,
				Score: r.GetScore(),
			})
			pw.Write(append(j, '\n'))
		}
	}
}

func (b *App) getTwLogEyeClient(src *LogSource) (api.TWLogEyeServiceClient, *grpc.ClientConn, error) {
	server := src.Server
	port := 8081
	if u, err := url.Parse(src.Server); err == nil && u.Hostname() != "" {
		server = u.Hostname()
		if u.Port() != "" {
			if p, err := strconv.Atoi(u.Port()); err == nil {
				port = p
			}
		}
	} else if strings.Contains(server, ":") {
		h, p, err := net.SplitHostPort(server)
		if err == nil {
			server = h
			if pn, err := strconv.Atoi(p); err == nil {
				port = pn
			}
		}
	}

	address := fmt.Sprintf("%s:%d", server, port)
	var conn *grpc.ClientConn
	var err error

	if src.CACert == "" && !src.TLS {
		conn, err = grpc.NewClient(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect: %w", err)
		}
	} else {
		if src.ClientCert != "" && src.ClientKey != "" {
			cert, err := tls.LoadX509KeyPair(src.ClientCert, src.ClientKey)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load client cert: %w", err)
			}
			tlsConfig := &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: src.InsecureSkip,
			}
			if src.CACert != "" {
				ca := x509.NewCertPool()
				caBytes, err := os.ReadFile(src.CACert)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to read ca cert: %w", err)
				}
				if ok := ca.AppendCertsFromPEM(caBytes); !ok {
					return nil, nil, fmt.Errorf("failed to parse ca cert %q", src.CACert)
				}
				tlsConfig.RootCAs = ca
			}
			conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to connect: %w", err)
			}
		} else if src.CACert != "" {
			creds, err := credentials.NewClientTLSFromFile(src.CACert, "")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load credentials: %w", err)
			}
			conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(creds))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to connect: %w", err)
			}
		} else {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: src.InsecureSkip,
			}
			conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to connect: %w", err)
			}
		}
	}
	return api.NewTWLogEyeServiceClient(conn), conn, nil
}

func getTimeStr(t int64) string {
	return time.Unix(0, t).Format(time.RFC3339Nano)
}
