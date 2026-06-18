package bazarr

type Version struct {
	Data struct {
		BazarrVersion string `json:"bazarr_version"`
	} `json:"data"`
}

type SyncParams struct {
	Path string `json:"path"`
	Id   int    `json:"id"`
	Action string `json:"action"`
	Lang string `json:"language"`
	Type string `json:"type"`
	Gss string `json:"gss"`
	NoFramerateFix string `json:"no_fix_framerate"`
}

type Subtitle struct {
	Path    string `json:"path"`
	Code2   string `json:"code2"`
	FileSize int   `json:"file_size"`
}

type Movie struct {
	Title string `json:"title"`
	Monitored bool `json:"monitored"`
	RadarrId int `json:"radarrId"`
	ImdbId string `json:"imdbId"`
	Subtitles []Subtitle `json:"subtitles"`
}

type MoviesInfo struct {
	Data []Movie `json:"data"`
}

type Show struct {
	Title string `json:"title"`
	Monitored bool `json:"monitored"`
	SonarrId int `json:"sonarrSeriesId"`
	ImdbId    string `json:"imdbId"`
}

type ShowsInfo struct {
	Data []Show `json:"data"`
}

type Episode struct {
	Title           string     `json:"title"`
	Monitored       bool       `json:"monitored"`
	SonarrEpId      int        `json:"sonarrEpisodeId"`
	SeasonNumber    int        `json:"season"`
	EpisodeNumber   int        `json:"episode"`
	Subtitles       []Subtitle `json:"subtitles"`
}

type EpisodesInfo struct {
	Data []Episode `json:"data"`
}
