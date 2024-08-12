package addr

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"log/slog"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/covrom/geoip/internal/cdn"
)

var (
	ErrNotFound = errors.New("target not found")

	mu      sync.RWMutex
	ipItems IpItems
)

type IpItem struct {
	Start   *big.Int
	End     *big.Int
	Country string
}

type IpItems []IpItem

func New() IpItems {
	var ip_items IpItems
	ip_items = append(ip_items, parseCsv(cdn.Ipv4Csv())...)
	ip_items = append(ip_items, parseCsv(cdn.Ipv6Csv())...)

	sort.Slice(ip_items, func(i, j int) bool {
		return ip_items[i].Start.Cmp(ip_items[j].Start) < 0
	})
	return ip_items
}

func (array IpItems) Search(target *big.Int, lowIndex int, highIndex int) (int, error) {
	if highIndex < lowIndex || len(array) == 0 {
		return -1, ErrNotFound
	}
	mid := int(lowIndex + (highIndex-lowIndex)/2)
	if array[mid].Start.Cmp(target) > 0 {
		return array.Search(target, lowIndex, mid-1)
	} else if array[mid].End.Cmp(target) < 0 {
		return array.Search(target, mid+1, highIndex)
	} else {
		return mid, nil
	}
}

func parseCsv(csvFile []byte) []IpItem {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("parseCsv error", "panic", r)
		}
	}()
	var items IpItems
	f := bytes.NewReader(csvFile)
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		start := new(big.Int)
		start, _ = start.SetString(record[0], 10)
		end := new(big.Int)
		end, _ = end.SetString(record[1], 10)
		items = append(items, IpItem{start, end, record[2]})
	}

	return items
}

func RegularUpdate(ctx context.Context, d time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := cdn.Update(); err != nil {
		slog.Error("RegularUpdate error", "err", err)
	} else {
		mu.Lock()
		ipItems = New()
		ln := len(ipItems)
		mu.Unlock()
		slog.Info("RegularUpdate: updated", "ip_count", ln)
	}

	tck := time.NewTicker(d)
	defer tck.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tck.C:
			if err := cdn.Update(); err != nil {
				slog.Error("RegularUpdate error", "err", err)
			} else {
				mu.Lock()
				ipItems = New()
				ln := len(ipItems)
				mu.Unlock()
				slog.Info("RegularUpdate: updated", "ip_count", ln)
			}
		}
	}
}

func Current() IpItems {
	mu.RLock()
	defer mu.RUnlock()

	return ipItems
}
