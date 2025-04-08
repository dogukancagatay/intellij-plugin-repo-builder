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

type LocalPlugin struct {
	ID          string `yaml:"id"`
	Version     string `yaml:"version"`
	Since       string `yaml:"since"`
	Until       string `yaml:"until"`
	File        string `yaml:"file"`
	Name        string `yaml:"name"`
	Vendor      string `yaml:"vendor"`
	VendorEmail string `yaml:"vendorEmail"`
	VendorUrl   string `yaml:"vendorUrl"`
	Description string `yaml:"description"`
}

type Config struct {
	ServerUrl    string        `yaml:"serverUrl"`
	BindIp       string        `yaml:"bindIp"`
	Port         string        `yaml:"port"`
	Dir          string        `yaml:"dir"`
	Plugins      []string      `yaml:"plugins"`
	LocalPlugins []LocalPlugin `yaml:"localPlugins"`
}

type Plugin struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	XMLID       string `json:"xmlId"`
	Description string `json:"description"`
}

type PluginRelease struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
	Since   string `json:"since"`
	Until   string `json:"until"`
	File    string `json:"file"`
}

type ReleaseDTO struct {
	ID      int    `json:"id"`
	File    string `json:"file"`
	Since   string `json:"since"`
	Until   string `json:"until"`
	Version string `json:"version"`
}

type PluginDTO struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	XMLID       string     `json:"xmlId"`
	Description string     `json:"description"`
	Vendor      string     `json:"vendor"`
	VendorEmail string     `json:"vendorEmail"`
	VendorUrl   string     `json:"vendorUrl"`
	Release     ReleaseDTO `json:"releases"`
}

func readConfig(filepath string) Config {
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
	httpClient := http.Client{Timeout: time.Second * 2}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	defer res.Body.Close()
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	return body
}

func getPlugin(pluginId string) Plugin {
	url := PluginApiUrlPrefix + "/" + pluginId
	jsonByteArr := httpGetRequest(url)
	plugin := Plugin{}
	err := json.Unmarshal(jsonByteArr, &plugin)
	if err != nil {
		log.Fatal(err)
	}
	return plugin
}

func getPluginReleases(pluginId string) []PluginRelease {
	url := PluginReleaseUrlPrefix + "/" + pluginId + "/" + PluginReleaseUrlSuffix
	jsonByteArr := httpGetRequest(url)
	var releases []PluginRelease
	err := json.Unmarshal(jsonByteArr, &releases)
	if err != nil {
		log.Fatal(err)
	}
	return releases
}

func downloadFile(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dirPath := filepath.Dir(filePath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			_ = os.MkdirAll(dirPath, 0755)
		}
		out, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		return err
	}
	return nil
}

func processPlugin(pluginId string, pluginMap map[string]PluginDTO, outputDir string) error {
	plugin := getPlugin(pluginId)
	releases := getPluginReleases(pluginId)
	sort.Slice(releases, func(i, j int) bool { return releases[i].Version > releases[j].Version })
	r := releases[0]
	url := IntellijDownloadUrlPrefix + "/" + r.File
	err := downloadFile(filepath.Join(outputDir, PluginDownloadDir, r.File), url)
	if err != nil {
		log.Fatalf("Failed to download plugin %s: %v", pluginId, err)
		return err
	}
	pluginMap[plugin.XMLID] = PluginDTO{
		ID:          plugin.ID,
		Name:        plugin.Name,
		XMLID:       plugin.XMLID,
		Description: plugin.Description,
		Release: ReleaseDTO{
			ID:      r.ID,
			File:    r.File,
			Version: r.Version,
			Since:   r.Since,
			Until:   r.Until,
		},
	}
	return nil
}

func copyLocalPlugin(p LocalPlugin, destDir string) error {
	destPath := filepath.Join(destDir, PluginDownloadDir, filepath.Base(p.File))
	input, err := os.Open(p.File)
	if err != nil {
		return err
	}
	defer input.Close()
	os.MkdirAll(filepath.Dir(destPath), 0755)
	output, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}

func writeLineListFile(filepath string, lineList []string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	datawriter := bufio.NewWriter(file)
	for _, data := range lineList {
		_, _ = datawriter.WriteString(data)
	}
	_ = datawriter.Flush()
	_ = file.Close()
}

func buildRepository(serverUrl string, pluginList []string, localPlugins []LocalPlugin, outputDir string) error {
	pluginMap := map[string]PluginDTO{}
	for _, pluginId := range pluginList {
		err := processPlugin(pluginId, pluginMap, outputDir)
		if err != nil {
			log.Fatalf("Failed to process plugin %s: %v", pluginId, err)
			return err
		}
	}
	for _, p := range localPlugins {
		err := copyLocalPlugin(p, outputDir)
		if err != nil {
			log.Fatalf("Failed to copy local plugin %s: %v", p.ID, err)
		}
		pluginMap[p.ID] = PluginDTO{
			ID:          0,
			Name:        p.Name,
			XMLID:       p.ID,
			Description: p.Description,
			Vendor:      p.Vendor,
			VendorEmail: p.VendorEmail,
			VendorUrl:   p.VendorUrl,
			Release: ReleaseDTO{
				File:    filepath.Base(p.File),
				Version: p.Version,
				Since:   p.Since,
				Until:   p.Until,
			},
		}
	}
	fileContent := []string{"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<plugins>\n"}
	for _, plugin := range pluginMap {
		fileContent = append(fileContent,
			fmt.Sprintf("\t<plugin id=\"%s\" url=\"%s/%s/%s\" version=\"%s\">\n",
				plugin.XMLID, serverUrl, PluginDownloadDir, plugin.Release.File, plugin.Release.Version))
		fileContent = append(fileContent,
			fmt.Sprintf("\t\t<name>%s</name>\n", plugin.Name))
		fileContent = append(fileContent,
			fmt.Sprintf("\t\t<vendor email=\"%s\" url=\"%s\"> %s </vendor>\n",
				plugin.VendorEmail, plugin.VendorUrl, plugin.Vendor))
		fileContent = append(fileContent,
			fmt.Sprintf("\t\t<description><![CDATA[ %s ]]></description>\n", plugin.Description))
		fileContent = append(fileContent,
			fmt.Sprintf("\t\t<idea-version since-build=\"%s\" until-build=\"%s\" />\n",
				plugin.Release.Since, plugin.Release.Until))
		fileContent = append(fileContent, "\t</plugin>\n")
	}
	fileContent = append(fileContent, "</plugins>\n")
	writeLineListFile(filepath.Join(outputDir, "updatePlugins.xml"), fileContent)
	log.Printf("📄 XML 路径: %s", filepath.Join(outputDir, "updatePlugins.xml"))
	return nil
}

func startHttpServer(host string, port string, dir string) {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)
	log.Println("Started listening on " + host + ":" + port + " ...")
	err := http.ListenAndServe(host+":"+port, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	serveHttp := flag.Bool("serve", false, "Start HTTP server")
	buildRepo := flag.Bool("build", false, "Build repository")
	configFile := flag.String("config", "config.yaml", "Config file")
	flag.Parse()
	config := readConfig(*configFile)
	if *serveHttp {
		startHttpServer(config.BindIp, config.Port, config.Dir)
	} else if *buildRepo {
		if len(config.Plugins) == 0 && len(config.LocalPlugins) == 0 {
			log.Fatalln("No plugins configured to build repository.")
		}
		err := buildRepository(config.ServerUrl, config.Plugins, config.LocalPlugins, config.Dir)
		if err != nil {
			log.Fatalf("Failed to build repository: %v\n", err)
			return
		}
	} else {
		fmt.Println("You need to specify one of `-build` or `-serve` arguments.")
	}
}
