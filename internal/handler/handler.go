package handler

import (
	"embed"
	"encoding/binary"
	"encoding/json"
	"expvar"
	"math/big"
	"net"
	"net/http"

	"github.com/covrom/geoip/internal/addr"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

type Response struct {
	Err     string `json:"error,omitempty"`
	IP      string `json:"ip,omitempty"`
	Country string `json:"country,omitempty"`
}

func New(static embed.FS) *http.ServeMux {
	mux := http.NewServeMux()

	fs := statigz.FileServer(static, brotli.AddEncoding, statigz.FSPrefix("web"))
	mux.Handle("GET /", fs)

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

	mux.Handle("GET /metrics", expvar.Handler())

	return mux
}
