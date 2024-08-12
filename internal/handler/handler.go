package handler

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net"
	"net/http"

	"github.com/covrom/geoip/internal/addr"
)

type Response struct {
	Err     string `json:"error,omitempty"`
	IP      string `json:"ip,omitempty"`
	Country string `json:"country,omitempty"`
}

func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /getIpInfo/{addr}", func(w http.ResponseWriter, r *http.Request) {
		addr_param := r.PathValue("addr")
		items := addr.Current()
		enc := json.NewEncoder(w)

		addr := net.ParseIP(addr_param)
		if addr != nil {
			ip_num := big.NewInt(0)

			if addr.To4() != nil {
				ip_num = new(big.Int).SetUint64(uint64(binary.BigEndian.Uint32(addr.To4())))
			} else {
				ip_num.SetBytes(addr)
			}
			idx, _ := items.Search(ip_num, 0, len(items))
			if idx >= 0 && ip_num.Cmp(big.NewInt(0)) != 0 {
				enc.Encode(Response{
					IP:      addr_param,
					Country: items[idx].Country,
				})
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		enc.Encode(Response{
			IP:  addr_param,
			Err: "not found",
		})
	})

	return mux
}
