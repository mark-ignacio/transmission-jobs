package jobs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hekmon/transmissionrpc"
)

// ConnectToRemote creates an *transmissionrpc.Client.
func ConnectToRemote(settings TransmissionSettings) (*transmissionrpc.Client, error) {
	uri, err := url.Parse(settings.Host)
	if err != nil {
		return nil, fmt.Errorf("error parsing Host: %+v", err)
	}
	port, err := strconv.ParseUint(uri.Port(), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("error parsing port as uint16: %+v", err)
	}
	advancedConfig := &transmissionrpc.AdvancedConfig{
		Port:  uint16(port),
		HTTPS: (strings.ToLower(uri.Scheme) == "https"),
	}
	return transmissionrpc.New(
		uri.Hostname(),
		settings.Username,
		settings.Password,
		advancedConfig,
	)
}

const sonarrHistoryPageSize = 50

type sonarrHistoryResponse struct {
	Page         int
	PageSize     int
	TotalRecords int
	Records      []struct {
		SourceTitle string
		EventType   string
		Data        struct {
			DroppedPath  string
			ImportedPath string
		}
	}
}

// FetchSonarrDrops crawls Sonarr's history for filesystem drop paths, going back up to maxRecords items.
// If maxRecords <= 0, this exhausts all possbile history items.
func FetchSonarrDrops(settings SonarrSettings, maxRecords int) (paths map[string]bool, err error) {
	endpoint, err := url.Parse(settings.Host)
	if err != nil {
		err = fmt.Errorf("error parsing Sonarr URL '%s': %+v", settings.Host, err)
		return
	}
	paths = make(map[string]bool)
	endpoint.Path = "/api/history"
	query := endpoint.Query()
	query.Set("apikey", settings.APIKey)
	query.Set("sortkey", "date")
	query.Set("sortDir", "desc")
	query.Set("pageSize", strconv.FormatInt(sonarrHistoryPageSize, 10))
	for currentPage := 1; ; currentPage++ {
		query.Set("page", strconv.FormatInt(int64(currentPage), 10))
		endpoint.RawQuery = query.Encode()
		var (
			response     *http.Response
			responseBody []byte
			history      sonarrHistoryResponse
		)
		response, err = http.DefaultClient.Get(endpoint.String())
		if err != nil {
			err = fmt.Errorf("error getting page %d of sonarr history: %+v", currentPage, err)
			return
		}
		responseBody, err = ioutil.ReadAll(response.Body)
		if err != nil {
			err = fmt.Errorf("error reading response body: %+v", err)
			return
		}
		if response.StatusCode != http.StatusOK {
			err = fmt.Errorf("%d code getting page %d of sonarr history", response.StatusCode, currentPage)
			log.Printf("response: %s", responseBody)
			return
		}
		err = json.Unmarshal(responseBody, &history)
		if err != nil {
			err = fmt.Errorf("error unmarshalling response body as JSON: %+v", err)
			return
		}
		for _, record := range history.Records {
			if record.Data.DroppedPath == "" {
				continue
			}
			paths[record.Data.DroppedPath] = true
		}
		// terminate loop?
		nextMaxRecord := (currentPage + 1) * sonarrHistoryPageSize
		if nextMaxRecord > history.TotalRecords || (maxRecords > 0 && nextMaxRecord > maxRecords) {
			break
		}
	}
	return
}
