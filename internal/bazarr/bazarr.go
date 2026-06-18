package bazarr

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ajmandourah/bazarr-sync/internal/client"
	"github.com/ajmandourah/bazarr-sync/internal/config"
)

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

func QueryMovies(cfg config.Config) (MoviesInfo, error) {
	return queryJSON[MoviesInfo](cfg, "movies")
}

func QuerySeries(cfg config.Config) (ShowsInfo, error) {
	return queryJSON[ShowsInfo](cfg, "series")
}

func QueryEpisodes(cfg config.Config, seriesId int) (EpisodesInfo, error) {
	c := client.GetClient(cfg.ApiToken)
	fullUrl := fmt.Sprintf("%sepisodes?seriesid[]=%d", cfg.ApiUrl, seriesId)
	resp, err := c.Get(fullUrl)
	if err != nil {
		return EpisodesInfo{}, fmt.Errorf("connection error to episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return EpisodesInfo{}, fmt.Errorf("unexpected status %d from episodes endpoint", resp.StatusCode)
	}

	var data EpisodesInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return EpisodesInfo{}, fmt.Errorf("failed to decode episodes response: %w", err)
	}
	return data, nil
}

func GetSyncParams(mediaType string, id int, sub Subtitle) SyncParams {
	return SyncParams{
		Action:         "sync",
		Path:           sub.Path,
		Id:             id,
		Lang:           sub.Code2,
		Type:           mediaType,
		Gss:            "False",
		NoFramerateFix: "False",
	}
}

func Sync(cfg config.Config, params SyncParams) bool {
	c := client.GetSyncClient(cfg.ApiToken)
	apiUrl, _ := url.JoinPath(cfg.ApiUrl, "subtitles")

	_url, _ := url.Parse(apiUrl)
	query := _url.Query()
	query.Set("path", params.Path)
	query.Set("id", strconv.Itoa(params.Id))
	query.Set("action", "sync")
	query.Set("language", params.Lang)
	query.Set("type", params.Type)
	query.Set("gss", params.Gss)
	query.Set("no_fix_framerate", params.NoFramerateFix)
	_url.RawQuery = query.Encode()

	resp, err := c.Patch(_url.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 204
}

func CheckHealth(cfg config.Config) (string, error) {
	c := client.GetClient(cfg.ApiToken)
	apiUrl, _ := url.JoinPath(cfg.ApiUrl, "system/status")
	resp, err := c.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("connection error to Bazarr: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Bazarr returned status %d. Check your URL and API token", resp.StatusCode)
	}

	var data Version
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("error parsing Bazarr response: %w", err)
	}
	return data.Data.BazarrVersion, nil
}
