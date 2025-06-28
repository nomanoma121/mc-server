package server

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// AddVelocityServerConfig は velocity.toml に新しいサーバー設定を安全に追加します。
// 厳密な構造体の代わりに汎用的なマップを使い、パースエラーを回避します。
func AddVelocityServerConfig(filePath, serverName, address string) error {
	// velocity.tomlを読み込む
	content, err := os.ReadFile(filePath)
	if err != nil {
		// ファイルが存在しない場合も考慮し、新規作成フローへ
		if !os.IsNotExist(err) {
			return fmt.Errorf("velocity.tomlの読み込みに失敗しました: %w", err)
		}
		// ファイルがなければ空の内容として扱う
		content = []byte{}
	}

	// TOML全体を汎用マップにデコードする
	var config map[string]interface{}
	if err := toml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("TOMLのパースに失敗しました: %w", err)
	}

	// 'servers' セクションを取得または新規作成する
	var servers map[string]interface{}
	if serversRaw, exists := config["servers"]; exists {
		// 既存の 'servers' セクションをマップに型アサーション
		var ok bool
		servers, ok = serversRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'servers'セクションのデータ型が不正です（マップではありません）")
		}
	} else {
		// 'servers' セクションが存在しなければ新しく作成
		servers = make(map[string]interface{})
		config["servers"] = servers
	}

	// 新しいサーバーの情報をマップに追加
	servers[serverName] = address

	// 更新した設定をTOML形式にエンコード（マーシャル）
	updatedContent, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("TOMLへのエンコードに失敗しました: %w", err)
	}

	// ファイルに書き込む
	if err := os.WriteFile(filePath, updatedContent, 0644); err != nil {
		return fmt.Errorf("velocity.tomlへの書き込みに失敗しました: %w", err)
	}

	return nil
}
