# Intellij Idea Custom Plugin Repository Builder

Custom Repository Builder provides tools for creating a Custom Plugin Repository for Jetbrains IDEs. This tool is tested on IntelliJ IDEA.

For a Custom Plugin Repository, you will need plugin files, `updatePlugins.xml` file and serve these on a web server.

## Quickstart

First step is building your repository. Download your plugins (latest versions) for offline use and create `updatePlugins.xml` file.

```sh
./repo-builder -build
```

Serve repository with the HTTP server.

```sh
./repo-builder -serve
```

## Configuration

Configuration options can be specified in `config.yml`.

```go
type Config struct {
	ServerUrl string   `yaml:"serverUrl"`
	BindIp    string   `yaml:"bindIp"`
	Port      string   `yaml:"port"`
	Dir       string   `yaml:"dir"`
	Plugins   []string `yaml:"plugins"`
}
```

- `serverUrl`: URL of the server that will serve the repository. (Default: `http://localhost:3000`)
- `bindIp`: IP address to bind HTTP Server to (Default: `0.0.0.0`)
- `port`: Port number of the HTTP Server (Default: `3000`)
- `dir`: Directory to build repository on and server from HTTP server from (Default: `out`)
- `plugins`: List of plugin IDs you want to add to repository. (*Required*)


Note that, `plugins` list should only contain IDs of the plugins which can be obtained from its URL.

For example, the following is the URL of the IdeaVim plugin, its plugin ID is *164*.

```
https://plugins.jetbrains.com/plugin/164-ideavim
```

You can find a sample list inside `config.yml`.


## IntelliJ IDE Setup

On your IntelliJ IDE, you need to set your Custom Plugin Repository URL according to [Jetbrains documentation](https://www.jetbrains.com/help/idea/managing-plugins.html#repos).

Afterwards, you will see the plugins in your repository on Marketplace tab.


## TODO

- Serving `updatePlugins.xml` according to version of Intellj Idea.
- Test and make it work on other repositories.

### References
- https://plugins.jetbrains.com/docs/intellij/update-plugins-format.html#format-of-updatepluginsxml-file
