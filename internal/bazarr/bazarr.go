package bazarr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ajmandourah/bazarr-sync/internal/client"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"github.com/pterm/pterm"
)

var cfg config.Config

// queryJSON makes a GET request to the Bazarr API and decodes the JSON response
// into the provided target. Returns an error if the request fails or the response
// status is not 200.
func queryJSON[T any](cfg config.Config, endpoint string) (T, error) {
	c := client.GetClient(cfg.ApiToken)
	apiUrl, _ := url.JoinPath(cfg.ApiUrl, endpoint)
	resp, err := c.Get(apiUrl)
	if err != nil {
		return *new(T), fmt.Errorf("connection error to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return *new(T), fmt.Errorf("unexpected status %d from %s", resp.StatusCode, endpoint)
	}

	var data T
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return *new(T), fmt.Errorf("failed to decode %s response: %w", endpoint, err)
	}
	return data, nil
}

func QueryMovies(cfg config.Config) (movies_info, error) {
	return queryJSON[movies_info](cfg, "movies")
}

func QuerySeries(cfg config.Config) (shows_info, error){
	return queryJSON[shows_info](cfg, "series")
}

func QueryEpisodes(cfg config.Config, seriesId int) (episodes_info,error){
	c := client.GetClient(cfg.ApiToken)
	fullUrl := fmt.Sprintf("%sepisodes?seriesid[]=%d", cfg.ApiUrl, seriesId)
	resp, err := c.Get(fullUrl)
	if err != nil {
		return episodes_info{}, fmt.Errorf("connection error to episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return episodes_info{}, fmt.Errorf("unexpected status %d from episodes endpoint", resp.StatusCode)
	}

	var data episodes_info
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return episodes_info{}, fmt.Errorf("failed to decode episodes response: %w", err)
	}
	return data, nil
}

func GetSyncParams(_type string, id int, subtitleInfo subtitle_info) Sync_params{
	var params Sync_params
	params.Action = "sync"
	params.Path = subtitleInfo.Path
	params.Id = id
	params.Lang = subtitleInfo.Code2
	params.Type = _type
	params.Gss = "False"	
	params.No_framerate_fix = "False"
	return params
}

func Sync(cfg config.Config, params Sync_params) bool {
	c := client.GetSyncClient(cfg.ApiToken)
	apiUrl , _ := url.JoinPath(cfg.ApiUrl, "subtitles")

	_url,_ := url.Parse(apiUrl)
	queryUrl := _url.Query()
	queryUrl.Set("path", params.Path)
	queryUrl.Set("id", strconv.Itoa(params.Id))
	queryUrl.Set("action", "sync")
	queryUrl.Set("language", params.Lang)
	queryUrl.Set("type", params.Type)
	queryUrl.Set("gss", params.Gss)
	queryUrl.Set("no_fix_framerate", params.No_framerate_fix)
	_url.RawQuery = queryUrl.Encode()
	resp, err := c.Patch(_url.String())
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {	
		return false
	}
	return true
}

func HealthCheck(cfg config.Config) {
	c := client.GetClient(cfg.ApiToken)
	apiUrl , _ := url.JoinPath(cfg.ApiUrl, "system/status")
	resp, err := c.Get(apiUrl)
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: Bazarr returned status", resp.StatusCode, ". Check your URL and API token.")
		os.Exit(1)
	}

	var data version
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing Bazarr response: ", err)
		os.Exit(1)
	}
	fmt.Println("Bazarr version: ", pterm.LightBlue(data.Data.Bazarr_version))
}
