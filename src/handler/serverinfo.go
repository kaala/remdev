package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
)

// ServerInfo holds server identification info.
type ServerInfo struct {
	Hostname string   `json:"hostname"`
	Addrs    []string `json:"addrs"`
}

// ServerInfoHandler returns server hostname and non-loopback IPs.
func ServerInfoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := ServerInfo{}

		hostname, _ := os.Hostname()
		info.Hostname = hostname

		ifaces, _ := net.Interfaces()
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 {
				continue
			}
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					continue
				}
				if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
					continue
				}
				if ip4 := ip.To4(); ip4 != nil {
					info.Addrs = append(info.Addrs, ip4.String())
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	})
}
