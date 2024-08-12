package cdn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	githubRepo = "sapics/ip-location-db"
	ipv4file   = "geo-whois-asn-country/geo-whois-asn-country-ipv4-num.csv"
	ipv6file   = "geo-asn-country/geo-asn-country-ipv6-num.csv"
)

var (
	url4     = fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s", githubRepo, ipv4file)
	url4hash = fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", githubRepo, ipv4file)

	url6     = fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s", githubRepo, ipv6file)
	url6hash = fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", githubRepo, ipv6file)

	mu sync.RWMutex

	ipv4csv,
	ipv6csv []byte

	hcli = &http.Client{
		Timeout: 10 * time.Minute,
	}
)

func getSHA256(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash
}

func getRemoteSHA256(urlFile string) (string, error) {
	resp, err := hcli.Get(urlFile)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type GitHubResponse struct {
		Sha string `json:"sha"`
	}

	var gitHubResponse GitHubResponse
	err = json.NewDecoder(resp.Body).Decode(&gitHubResponse)
	if err != nil {
		return "", err
	}

	return gitHubResponse.Sha, nil
}

func Update() error {
	csv4 := Ipv4Csv()
	if len(csv4) > 0 {
		localSha := getSHA256(csv4)
		remoteSha, err := getRemoteSHA256(url4hash)
		if err != nil {
			return fmt.Errorf("getRemoteSHA256(url4hash) eror: %w", err)
		}
		if localSha != remoteSha {
			if err := downloadFile(&ipv4csv, url4); err != nil {
				return fmt.Errorf("downloadFile(&ipv4csv, url4) error: %w", err)
			}
		}
	} else {
		if err := downloadFile(&ipv4csv, url4); err != nil {
			return fmt.Errorf("downloadFile(&ipv4csv, url4) error: %w", err)
		}
	}

	csv6 := Ipv6Csv()
	if len(csv6) > 0 {
		localSha := getSHA256(csv6)
		remoteSha, err := getRemoteSHA256(url6hash)
		if err != nil {
			return fmt.Errorf("getRemoteSHA256(url6hash) eror: %w", err)
		}
		if localSha != remoteSha {
			if err := downloadFile(&ipv6csv, url6); err != nil {
				return fmt.Errorf("downloadFile(&ipv6csv, url6) error: %w", err)
			}
		}
	} else {
		if err := downloadFile(&ipv6csv, url6); err != nil {
			return fmt.Errorf("downloadFile(&ipv6csv, url6) error: %w", err)
		}
	}

	return nil
}

func downloadFile(dest *[]byte, url string) error {
	resp, err := hcli.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	mu.Lock()
	*dest = b
	mu.Unlock()

	return err
}

func Ipv4Csv() []byte {
	mu.RLock()
	defer mu.RUnlock()

	return ipv4csv
}

func Ipv6Csv() []byte {
	mu.RLock()
	defer mu.RUnlock()

	return ipv6csv
}
