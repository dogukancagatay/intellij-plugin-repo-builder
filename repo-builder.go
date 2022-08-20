package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const PluginDownloadDir string = "files"
const IntellijDownloadUrlPrefix = "https://plugins.jetbrains.com/files"
const PluginApiUrlPrefix = "https://plugins.jetbrains.com/api/plugins"
const PluginReleaseUrlPrefix = "https://plugins.jetbrains.com/api/plugins"
const PluginReleaseUrlSuffix = "updates?channel=&size=8"

type Plugin struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Approve       bool   `json:"approve"`
	XMLID         string `json:"xmlId"`
	Description   string `json:"description"`
	CustomIdeList bool   `json:"customIdeList"`
	Preview       string `json:"preview"`
	DocText       string `json:"docText"`
	Cdate         int64  `json:"cdate"`
	Family        string `json:"family"`
	Downloads     int    `json:"downloads"`
	Vendor        struct {
		Name       string `json:"name"`
		URL        string `json:"url"`
		IsVerified bool   `json:"isVerified"`
	} `json:"vendor"`
	Urls struct {
		URL           string `json:"url"`
		ForumURL      string `json:"forumUrl"`
		LicenseURL    string `json:"licenseUrl"`
		BugtrackerURL string `json:"bugtrackerUrl"`
		DocURL        string `json:"docUrl"`
		SourceCodeURL string `json:"sourceCodeUrl"`
	} `json:"urls"`
	Tags []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Privileged bool   `json:"privileged"`
		Searchable bool   `json:"searchable"`
		Link       string `json:"link"`
	} `json:"tags"`
	HasUnapprovedUpdate bool   `json:"hasUnapprovedUpdate"`
	ReadyForSale        bool   `json:"readyForSale"`
	Icon                string `json:"icon"`
}

type PluginRelease struct {
	ID                              int    `json:"id"`
	Link                            string `json:"link"`
	Version                         string `json:"version"`
	Approve                         bool   `json:"approve"`
	Listed                          bool   `json:"listed"`
	RecalculateCompatibilityAllowed bool   `json:"recalculateCompatibilityAllowed"`
	Cdate                           string `json:"cdate"`
	File                            string `json:"file"`
	Notes                           string `json:"notes"`
	Since                           string `json:"since"`
	Until                           string `json:"until"`
	SinceUntil                      string `json:"sinceUntil"`
	Channel                         string `json:"channel"`
	Size                            int    `json:"size"`
	Downloads                       int    `json:"downloads"`
	PluginID                        int    `json:"pluginId"`
	CompatibleVersions              struct {
		IdeaEducational    string `json:"IDEA_EDUCATIONAL"`
		Appcode            string `json:"APPCODE"`
		Phpstorm           string `json:"PHPSTORM"`
		Rider              string `json:"RIDER"`
		Clion              string `json:"CLION"`
		PycharmCommunity   string `json:"PYCHARM_COMMUNITY"`
		AndroidStudio      string `json:"ANDROID_STUDIO"`
		IdeaCommunity      string `json:"IDEA_COMMUNITY"`
		Rubymine           string `json:"RUBYMINE"`
		Dataspell          string `json:"DATASPELL"`
		PycharmEducational string `json:"PYCHARM_EDUCATIONAL"`
		Webstorm           string `json:"WEBSTORM"`
		Mps                string `json:"MPS"`
		Idea               string `json:"IDEA"`
		Dbe                string `json:"DBE"`
		Pycharm            string `json:"PYCHARM"`
		Goland             string `json:"GOLAND"`
	} `json:"compatibleVersions"`
	Author struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Link     string `json:"link"`
		HubLogin string `json:"hubLogin"`
		Icon     string `json:"icon"`
	} `json:"author"`
	Modules []interface{} `json:"modules"`
}

type ReleaseDTO struct {
	ID      int    `json:"id"`
	File    string `json:"file"`
	Since   string `json:"since"`
	Until   string `json:"until"`
	Version string `json:"version"`
}

type PluginDTO struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	XMLID   string     `json:"xmlId"`
	Release ReleaseDTO `json:"releases"`
}

type Config struct {
	ServerUrl string   `yaml:"serverUrl"`
	BindIp    string   `yaml:"bindIp"`
	Port      string   `yaml:"port"`
	Dir       string   `yaml:"dir"`
	Plugins   []string `yaml:"plugins"`
}

func readConfig(filepath string) Config {
	log.Printf("Read config from %s\n", filepath)
	yfile, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	data := Config{
		ServerUrl: "http://localhost:3000",
		BindIp:    "0.0.0.0",
		Port:      "3000",
		Dir:       "out",
	}

	err2 := yaml.Unmarshal(yfile, &data)
	if err2 != nil {
		log.Fatal(err2)
	}

	return data
}

func httpGetRequest(url string) []byte {
	httpClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return body[:]
}

func getPlugin(pluginId string) Plugin {
	url := PluginApiUrlPrefix + "/" + pluginId

	jsonByteArr := httpGetRequest(url)
	plugin := Plugin{}
	jsonErr := json.Unmarshal(jsonByteArr, &plugin)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return plugin
}

func getPluginReleases(pluginId string) []PluginRelease {
	url := PluginReleaseUrlPrefix + "/" + pluginId + "/" + PluginReleaseUrlSuffix

	jsonByteArr := httpGetRequest(url)
	var pluginReleaseArr []PluginRelease
	jsonErr := json.Unmarshal(jsonByteArr, &pluginReleaseArr)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return pluginReleaseArr
}

func downloadFile(filePath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error accessing plugin API: %s", err)
		return err
	}
	defer resp.Body.Close()

	// download if if file doesn't exist
	if _, err = os.Stat(filePath); os.IsNotExist(err) {

		// fp := filepath.Join(filePath)
		dirPath := filepath.Dir(filePath)

		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, 0755)
			if err != nil {
				log.Fatalf("Error creating download directories: %s", err)
				return err
			}
		}

		// Create the file
		out, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("Error creating file: %s", err)
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatalf("Could not write body to file %s", filePath)
			return err
		}
		log.Printf("Downloaded %s to %s", url, filePath)
	} else {
		log.Printf("File is already exists: %s", filePath)
	}

	return nil
}

func processPlugin(pluginId string, sinceMap map[string]PluginDTO, outputDir string) {

	downloadDirPath := outputDir + "/" + PluginDownloadDir

	plugin := getPlugin(pluginId)
	pluginReleases := getPluginReleases(pluginId)

	sort.Slice(pluginReleases, func(i, j int) bool {
		return pluginReleases[i].Version > pluginReleases[j].Version
	})

	r := pluginReleases[0]
	downloadUrl := IntellijDownloadUrlPrefix + "/" + r.File

	log.Printf("Will download '%s' (%s) (%s) from %s\n", plugin.Name, plugin.XMLID, r.Version, downloadUrl)

	// Create Release DTO
	release := ReleaseDTO{}
	release.ID = r.ID
	release.File = r.File
	release.Since = r.Since
	release.Until = r.Until
	release.Version = r.Version

	// Create pluginDTO
	pluginDto := PluginDTO{}
	pluginDto.ID = plugin.ID
	pluginDto.Name = plugin.Name
	pluginDto.XMLID = plugin.XMLID
	pluginDto.Release = release

	sinceMap[plugin.XMLID] = pluginDto
	err := downloadFile(downloadDirPath+"/"+r.File, downloadUrl)
	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
		return
	}

}

func writeLineListFile(filepath string, lineList []string) {

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)

	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
		return
	}

	datawriter := bufio.NewWriter(file)

	for _, data := range lineList {
		_, _ = datawriter.WriteString(data)
	}

	err = datawriter.Flush()
	if err != nil {
		log.Fatalf("Failed flushing file: %s", err)
		return
	}

	err = file.Close()
	if err != nil {
		log.Fatalf("Failed closing file: %s", err)
		return
	}
}

func buildRepository(serverUrl string, pluginList []string, outputDir string) {
	// Download Plugins

	log.Println("Obtain plugin data and download plugins")
	var sinceMap = map[string]PluginDTO{}

	for _, pluginId := range pluginList {
		processPlugin(pluginId, sinceMap, outputDir)
	}

	// Prepare file content
	log.Println("Prepare updatePlugins.xml file content")
	fileContent := []string{"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<plugins>\n"}

	for _, plugin := range sinceMap {
		fileContent = append(fileContent,
			fmt.Sprintf("\t<plugin id=\"%s\" url=\"%s/%s/%s\" version=\"%s\">\n",
				plugin.XMLID,
				serverUrl,
				PluginDownloadDir,
				plugin.Release.File,
				plugin.Release.Version,
			),
		)

		fileContent = append(fileContent,
			fmt.Sprintf("\t\t<idea-version since-build=\"%s\" until-build=\"%s\" />\n",
				plugin.Release.Since,
				plugin.Release.Until,
			),
		)
		fileContent = append(fileContent, "\t</plugin>\n")
	}

	fileContent = append(fileContent, "</plugins>\n")

	// Write to updatePlugins.xml file
	log.Println("Write updatePlugins.xml file")
	writeLineListFile(outputDir+"/updatePlugins.xml", fileContent)
}

func startHttpServer(host string, port string, dir string) {

	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	log.Println("Started listening on " + host + ":" + port + " ...")
	err := http.ListenAndServe(host+":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	// Arguments
	serveHttp := flag.Bool("serve", false, "Start HTTP server")
	buildRepo := flag.Bool("build", false, "Build repository (Requires internet)")
	configFile := flag.String("config", "config.yaml", "Config file")
	flag.Parse()

	// Read config
	config := readConfig(*configFile)

	// Run according to mode
	if *serveHttp {
		startHttpServer(config.BindIp, config.Port, config.Dir)
	} else if *buildRepo {
		if len(config.Plugins) == 0 {
			log.Fatalln("Cannot build repository on empty plugin list.")
		}
		buildRepository(config.ServerUrl, config.Plugins, config.Dir)
	} else {
		fmt.Println("You need to specify one of `-build` or `-serve` arguments.")
		fmt.Println("Usage: repo-builder <arguments>")
	}

}
