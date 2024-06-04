package bazarr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/ajmandourah/bazarr-sync/internal/client"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"github.com/pterm/pterm"
)

var cfg config.Config

func QueryMovies(cfg config.Config) (movies_info, error) {	
	c := client.GetClient(cfg.ApiToken)
	url , _ := url.JoinPath(cfg.ApiUrl, "movies")
	resp, err := c.Get(url) 
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return movies_info{},err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: ", "Response status is not 200. Are you sure the address/port are correct?")
		return movies_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {	
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return movies_info{},err
	}
	var data movies_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return movies_info{},err
	}
	return data, nil
}

func QuerySeries(cfg config.Config) (shows_info, error){
	c := client.GetClient(cfg.ApiToken)
	url , _ := url.JoinPath(cfg.ApiUrl, "series")
	resp, err := c.Get(url) 
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return shows_info{},err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: ", "Response status is not 200. Are you sure the address/port are correct?")
		return shows_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {	
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return shows_info{},err
	}
	var data shows_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return shows_info{},err
	}
	return data, nil
}

func QueryEpisodes(cfg config.Config, seriesId int) (episodes_info,error){
	c := client.GetClient(cfg.ApiToken)
	u , _ := url.JoinPath(cfg.ApiUrl, "episodes")
	_url,_ := url.Parse(u)
	queryUrl := _url.Query()
	queryUrl.Set("seriesid[]", strconv.Itoa(seriesId))
	_url.RawQuery = queryUrl.Encode()
	
	resp, err := c.Get(_url.String()) 
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return episodes_info{},err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: ", "Response status is not 200. Are you sure the address/port are correct?")
		return episodes_info{}, errors.New("Error: Status code not 200")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {	
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
		return episodes_info{},err
	}
	var data episodes_info
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error in Unmarshaling json body", err)
		return episodes_info{},err
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
	c := client.GetClient(cfg.ApiToken)
	u , _ := url.JoinPath(cfg.ApiUrl, "subtitles")

	_url,_ := url.Parse(u)
	queryUrl := _url.Query()
	queryUrl.Set("path", params.Path)
	queryUrl.Set("id", strconv.Itoa(params.Id))
	queryUrl.Set("action", "sync")
	queryUrl.Set("language", params.Lang)
	queryUrl.Set("type", params.Type)
	_url.RawQuery = queryUrl.Encode()
	resp, err := c.Patch(_url.String())
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
		return false
	}
	if resp.StatusCode != 204 {	
		return false
	}
	return true

}

func HealthCheck(cfg config.Config) {
	c := client.GetClient(cfg.ApiToken)
	url , _ := url.JoinPath(cfg.ApiUrl, "system/status")
	resp, err := c.Get(url) 
	if err != nil{
		fmt.Fprintln(os.Stderr, "Connection Error: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Connection Error: ", "Response status is not 200. Are you sure the address/port are correct?")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {	
		fmt.Fprintln(os.Stderr, "Reading Url Error: ", err)
	}
	var data version
	json.Unmarshal(body,&data)
	fmt.Println("Bazarr version: ", pterm.LightBlue(data.Data.Bazarr_version))
	return
}
