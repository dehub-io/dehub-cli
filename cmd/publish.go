package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var publishDir string

var publishCmd = &cobra.Command{
	Use:   "publish [directory]",
	Short: "发布包",
	Long: `发布包到 dehub 仓库。

会执行以下步骤：
1. 读取 package.yaml 获取包名和版本
2. 打包源码文件
3. 计算 SHA256 校验值
4. 上传到服务端
5. 触发验证流程`,
	Example: `  dehub publish
  dehub publish ./my-package`,
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		
		adapter := getAdapter()
		
		// 检查登录状态，未登录则自动触发登录
		status, err := adapter.GetAuthStatus()
		if err != nil || !status.LoggedIn {
			fmt.Println("未登录，正在启动登录流程...")
			if err := adapter.Login(); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "登录失败: %v\n", err)
				os.Exit(1)
			}
			// 重新获取状态
			status, _ = adapter.GetAuthStatus()
		}
		
		fmt.Printf("发布者: %s\n\n", status.Username)
		
		if err := adapter.Publish(dir, cmd.OutOrStdout()); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "发布失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	publishCmd.Flags().StringVarP(&publishDir, "dir", "d", ".", "包目录")
	rootCmd.AddCommand(publishCmd)
}
