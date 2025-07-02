package cmd

import (
	"fmt"
	"mcctl/internal/server"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "新しいMinecraftサーバーを追加します",
	Run: func(cmd *cobra.Command, args []string) {
		// サーバー名入力
		prompt := promptui.Prompt{Label: "サーバー名"}
		name, err := prompt.Run()
		if err != nil {
			fmt.Println("キャンセルされました")
			return
		}

		// バージョン選択
		versions := []string{"vanilla", "forge", "fabric"}
		versionPrompt := promptui.Select{
			Label: "サーバーバージョンを選択してください",
			Items: versions,
		}
		_, version, err := versionPrompt.Run()
		if err != nil {
			fmt.Println("キャンセルされました")
			return
		}

		// アドレス入力
		addressPrompt := promptui.Prompt{Label: "サーバーのアドレス（例: myserver:25565）"}
		address, err := addressPrompt.Run()
		if err != nil {
			fmt.Println("キャンセルされました")
			return
		}

		// 管理用JSONファイルに保存
		s := server.Server{Name: name, Version: version, Address: address}
		err = server.SaveServerConfig("minecraft/servers.json", s)
		if err != nil {
			fmt.Printf("サーバーの保存に失敗しました: %v\n", err)
			return
		}

		// サーバーディレクトリとテンプレートファイルを作成
		err = server.CreateServerDirectory(name, version)
		if err != nil {
			fmt.Printf("サーバーディレクトリの作成に失敗しました: %v\n", err)
			return
		}

		// velocity.tomlに追加
		err = server.AddVelocityServerConfig("velocity/velocity.toml", name, address)
		if err != nil {
			fmt.Printf("Velocity設定更新失敗: %v\n", err)
			return
		}

		// minecraft/docker-compose.ymlに追加
		err = server.AddDockerComposeService("minecraft/docker-compose.yml", name, version)
		if err != nil {
			fmt.Printf("Docker Compose設定更新失敗: %v\n", err)
			return
		}

		fmt.Printf("サーバー %s (タイプ: %s, アドレス: %s) を追加しました\n", name, version, address)
		fmt.Printf("minecraft/docker-compose.ymlにサービス '%s' を追加しました\n", name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
