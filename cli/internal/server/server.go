package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ServerTypeInterface defines the interface for different server types
type ServerTypeInterface interface {
	GetEnvironment() []string
	GetVolumes(serverName string) []string
	GetTemplatePath() string
	GetSubdirectories() []string
	GetTemplateFiles() []string
}

// ServerTypeFactory manages server type creation
type ServerTypeFactory struct {
	types map[string]func() ServerTypeInterface
}

// globalFactory is the global instance of the server type factory
var globalFactory = &ServerTypeFactory{
	types: make(map[string]func() ServerTypeInterface),
}

// RegisterServerType registers a new server type with the factory
func RegisterServerType(name string, factory func() ServerTypeInterface) {
	globalFactory.types[strings.ToLower(name)] = factory
}

// GetServerType returns the appropriate ServerTypeInterface implementation
func GetServerType(serverType string) (ServerTypeInterface, error) {
	factory, exists := globalFactory.types[strings.ToLower(serverType)]
	if !exists {
		return nil, fmt.Errorf("サポートされていないサーバータイプ: %s", serverType)
	}
	return factory(), nil
}

// GetRegisteredTypes returns all registered server type names
func GetRegisteredTypes() []string {
	types := make([]string, 0, len(globalFactory.types))
	for name := range globalFactory.types {
		types = append(types, name)
	}
	return types
}

// Server は、管理用のJSONファイルに保存するサーバー情報の構造体です。
type Server struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Address string `json:"address"` // 例: "myserver:25565"
}

// SaveServerConfig は、管理用JSONファイル（例: servers.json）にサーバー情報を保存します。
// この関数は velocity.toml とは無関係で、問題なく動作します。
func SaveServerConfig(jsonPath string, s Server) error {
	var servers []Server

	data, err := os.ReadFile(jsonPath)
	// ファイルが存在する場合、既存のデータを読み込む
	if err == nil {
		if err := json.Unmarshal(data, &servers); err != nil {
			return fmt.Errorf("管理用JSONのパースに失敗しました: %w", err)
		}
		// ファイルが存在しないエラー以外は、予期せぬエラーとして返す
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("管理用JSONの読み込みに失敗しました: %w", err)
	}

	// 新しいサーバー情報をスライスに追加
	servers = append(servers, s)

	// 整形したJSON形式で書き出す
	updated, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONへのエンコードに失敗しました: %w", err)
	}

	return os.WriteFile(jsonPath, updated, 0644)
}

func AddVelocityServerConfig(tomlPath, serverName, address string) error {
	// 1. velocity.toml を読み込む
	content, err := os.ReadFile(tomlPath)
	if err != nil {
		// ファイルが存在しない場合はエラーとせず、空の内容として新規作成フローに進む
		if !os.IsNotExist(err) {
			return fmt.Errorf("velocity.tomlの読み込みに失敗しました: %w", err)
		}
		content = []byte{}
	}

	// 2. 汎用的なマップにデコードする
	var config map[string]interface{}
	if err := toml.Unmarshal(content, &config); err != nil {
		// このデコードが失敗する場合、TOMLの構文自体に問題がある可能性が高い
		return fmt.Errorf("TOMLのパースに失敗しました: %w", err)
	}

	// 3. 'servers' セクションを安全に取得・更新する
	var servers map[string]interface{}
	if serversRaw, exists := config["servers"]; exists {
		// 型アサーションでマップに変換
		servers, _ = serversRaw.(map[string]interface{})
	}
	if servers == nil { // セクションが存在しないか、型が違う場合は新規作成
		servers = make(map[string]interface{})
		config["servers"] = servers
	}
	servers[serverName] = address

	// 4. 'forced-hosts' セクションを安全に取得・更新する
	//    TOMLのキーはハイフンを含む "forced-hosts" なので注意
	var forcedHosts map[string]interface{}
	if forcedHostsRaw, exists := config["forced-hosts"]; exists {
		forcedHosts, _ = forcedHostsRaw.(map[string]interface{})
	}
	if forcedHosts == nil { // セクションが存在しないか、型が違う場合は新規作成
		forcedHosts = make(map[string]interface{})
		config["forced-hosts"] = forcedHosts
	}
	// 新しいforced-hostを追加
	forcedHosts[serverName+".example.com"] = []string{serverName}

	// 5. 更新したマップをTOML形式に変換してファイルに書き込む
	updatedContent, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("TOMLへのエンコードに失敗しました: %w", err)
	}

	if err := os.WriteFile(tomlPath, updatedContent, 0644); err != nil {
		return fmt.Errorf("velocity.tomlへの書き込みに失敗しました: %w", err)
	}

	return nil
}

// DockerComposeService represents a service in docker-compose.yml
type DockerComposeService struct {
	Build struct {
		Context    string `yaml:"context"`
		Dockerfile string `yaml:"dockerfile"`
	} `yaml:"build,omitempty"`
	Image         string      `yaml:"image,omitempty"`
	ContainerName string      `yaml:"container_name,omitempty"`
	Environment   interface{} `yaml:"environment,omitempty"` // 配列またはマップ形式に対応
	Volumes       []string    `yaml:"volumes,omitempty"`
	Networks      []string    `yaml:"networks,omitempty"`
	Restart       string      `yaml:"restart,omitempty"`
	TTY           bool        `yaml:"tty,omitempty"`
	StdinOpen     bool        `yaml:"stdin_open,omitempty"`
}

// DockerCompose represents the structure of docker-compose.yml
type DockerCompose struct {
	Version  string                          `yaml:"version"`
	Services map[string]DockerComposeService `yaml:"services"`
	Networks map[string]interface{}          `yaml:"networks,omitempty"`
	Volumes  map[string]interface{}          `yaml:"volumes,omitempty"`
}

// AddDockerComposeService adds a new Minecraft server service to docker-compose.yml
func AddDockerComposeService(dockerComposePath, serverName, serverType string) error {
	// Get the appropriate server type implementation
	serverTypeImpl, err := GetServerType(serverType)
	if err != nil {
		return err
	}

	// Read existing docker-compose.yml
	var compose DockerCompose

	data, err := os.ReadFile(dockerComposePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create basic docker-compose.yml structure if it doesn't exist
			compose = DockerCompose{
				Version:  "3.8",
				Services: make(map[string]DockerComposeService),
				Networks: map[string]interface{}{
					"home-network": map[string]interface{}{
						"external": true,
					},
				},
			}
		} else {
			return fmt.Errorf("docker-compose.ymlの読み込みに失敗しました: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &compose); err != nil {
			return fmt.Errorf("docker-compose.ymlのパースに失敗しました: %w", err)
		}
	}

	// Ensure services map exists
	if compose.Services == nil {
		compose.Services = make(map[string]DockerComposeService)
	}

	// Create new service using the interface
	newService := DockerComposeService{
		Build: struct {
			Context    string `yaml:"context"`
			Dockerfile string `yaml:"dockerfile"`
		}{
			Context:    serverTypeImpl.GetTemplatePath(),
			Dockerfile: "Dockerfile",
		},
		ContainerName: fmt.Sprintf("minecraft-%s-server", serverName),
		Environment:   serverTypeImpl.GetEnvironment(),
		Volumes:       serverTypeImpl.GetVolumes(serverName),
		Networks:      []string{"home-network"},
		Restart:       "unless-stopped",
		TTY:           true,
		StdinOpen:     true,
	}

	// Add new service to compose
	compose.Services[serverName] = newService

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(&compose)
	if err != nil {
		return fmt.Errorf("docker-compose.ymlのエンコードに失敗しました: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(dockerComposePath, updatedData, 0644); err != nil {
		return fmt.Errorf("docker-compose.ymlの書き込みに失敗しました: %w", err)
	}

	return nil
}

// CreateServerDirectory creates the server directory structure and copies template files
func CreateServerDirectory(serverName, serverType string) error {
	// Get the appropriate server type implementation
	serverTypeImpl, err := GetServerType(serverType)
	if err != nil {
		return err
	}

	serverDir := fmt.Sprintf("minecraft/servers/%s", serverName)
	templateDir := fmt.Sprintf("minecraft/template/%s", serverType)

	// Create server directory
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("サーバーディレクトリの作成に失敗しました: %w", err)
	}

	// Create subdirectories using the interface
	for _, subdir := range serverTypeImpl.GetSubdirectories() {
		if err := os.MkdirAll(fmt.Sprintf("%s/%s", serverDir, subdir), 0755); err != nil {
			return fmt.Errorf("サブディレクトリ %s の作成に失敗しました: %w", subdir, err)
		}
	}

	// Copy template files using the interface
	for _, file := range serverTypeImpl.GetTemplateFiles() {
		src := fmt.Sprintf("%s/%s", templateDir, file)
		dst := fmt.Sprintf("%s/%s", serverDir, file)

		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("テンプレートファイル %s のコピーに失敗しました: %w", file, err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// ForgeServerType implements ServerTypeInterface for Forge servers
type ForgeServerType struct{}

func (f *ForgeServerType) GetEnvironment() []string {
	return []string{
		"EULA=true",
		"TYPE=FORGE",
		"VERSION=1.18.2",
		"MEMORY=4G",
	}
}

func (f *ForgeServerType) GetVolumes(serverName string) []string {
	return []string{
		fmt.Sprintf("./servers/%s/world:/data/world", serverName),
		fmt.Sprintf("./servers/%s/mods:/data/mods", serverName),
		fmt.Sprintf("./servers/%s/config:/data/config", serverName),
		fmt.Sprintf("./servers/%s/ops.json:/data/ops.json", serverName),
		fmt.Sprintf("./servers/%s/server.properties:/data/server.properties", serverName),
		fmt.Sprintf("./servers/%s/whitelist.json:/data/whitelist.json", serverName),
	}
}

func (f *ForgeServerType) GetTemplatePath() string {
	return "./template/forge"
}

func (f *ForgeServerType) GetSubdirectories() []string {
	return []string{"world", "mods", "config"}
}

func (f *ForgeServerType) GetTemplateFiles() []string {
	return []string{"ops.json", "whitelist.json", "server.properties"}
}

// PaperServerType implements ServerTypeInterface for Paper servers
type PaperServerType struct{}

func (p *PaperServerType) GetEnvironment() []string {
	return []string{
		"EULA=true",
		"TYPE=PAPER",
		"VERSION=1.20.1",
		"MEMORY=4G",
	}
}

func (p *PaperServerType) GetVolumes(serverName string) []string {
	return []string{
		fmt.Sprintf("./servers/%s/world:/data/world", serverName),
		fmt.Sprintf("./servers/%s/plugins:/data/plugins", serverName),
		fmt.Sprintf("./servers/%s/ops.json:/data/ops.json", serverName),
		fmt.Sprintf("./servers/%s/paper-global.yml:/config/paper-global.yml", serverName),
		fmt.Sprintf("./servers/%s/server.properties:/data/server.properties", serverName),
		fmt.Sprintf("./servers/%s/whitelist.json:/data/whitelist.json", serverName),
	}
}

func (p *PaperServerType) GetTemplatePath() string {
	return "./template/paper"
}

func (p *PaperServerType) GetSubdirectories() []string {
	return []string{"world", "plugins"}
}

func (p *PaperServerType) GetTemplateFiles() []string {
	return []string{"ops.json", "whitelist.json", "server.properties", "paper-global.yml"}
}

// VanillaServerType implements ServerTypeInterface for Vanilla servers
type VanillaServerType struct{}

func (v *VanillaServerType) GetEnvironment() []string {
	return []string{
		"EULA=true",
		"TYPE=VANILLA",
		"VERSION=1.20.1",
		"MEMORY=2G",
	}
}

func (v *VanillaServerType) GetVolumes(serverName string) []string {
	return []string{
		fmt.Sprintf("./servers/%s/world:/data/world", serverName),
		fmt.Sprintf("./servers/%s/ops.json:/data/ops.json", serverName),
		fmt.Sprintf("./servers/%s/server.properties:/data/server.properties", serverName),
		fmt.Sprintf("./servers/%s/whitelist.json:/data/whitelist.json", serverName),
	}
}

func (v *VanillaServerType) GetTemplatePath() string {
	return "./template/vanilla"
}

func (v *VanillaServerType) GetSubdirectories() []string {
	return []string{"world"}
}

func (v *VanillaServerType) GetTemplateFiles() []string {
	return []string{"ops.json", "whitelist.json", "server.properties"}
}

// init registers all default server types
func init() {
	RegisterServerType("forge", func() ServerTypeInterface { return &ForgeServerType{} })
	RegisterServerType("paper", func() ServerTypeInterface { return &PaperServerType{} })
	RegisterServerType("vanilla", func() ServerTypeInterface { return &VanillaServerType{} })
}
