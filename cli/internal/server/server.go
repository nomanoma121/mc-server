package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

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
