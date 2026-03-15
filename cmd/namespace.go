package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "管理命名空间",
}

var namespaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出命名空间",
	Run: func(cmd *cobra.Command, args []string) {
		adapter := getAdapter()
		namespaces, err := adapter.ListNamespaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取命名空间列表失败: %v\n", err)
			os.Exit(1)
		}

		if len(namespaces) == 0 {
			fmt.Println("暂无命名空间")
			return
		}

		fmt.Println("命名空间列表:")
		for _, ns := range namespaces {
			fmt.Printf("  %s", ns.Name)
			if ns.Description != "" {
				fmt.Printf(" - %s", ns.Description)
			}
			fmt.Println()
		}
	},
}

var namespaceDesc string

var namespaceCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "申请命名空间",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		adapter := getAdapter()

		// 检查登录状态
		status, err := adapter.GetAuthStatus()
		if err != nil || !status.LoggedIn {
			fmt.Println("未登录，正在启动登录流程...")
			if err := adapter.Login(); err != nil {
				fmt.Fprintf(os.Stderr, "登录失败: %v\n", err)
				os.Exit(1)
			}
		}

		if _, err := adapter.CreateNamespace(name, namespaceDesc); err != nil {
			fmt.Fprintf(os.Stderr, "申请失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("命名空间 '%s' 申请已提交，等待审核\n", name)
	},
}

func init() {
	namespaceCreateCmd.Flags().StringVarP(&namespaceDesc, "description", "d", "", "命名空间描述")
	namespaceCmd.AddCommand(namespaceListCmd)
	namespaceCmd.AddCommand(namespaceCreateCmd)
	rootCmd.AddCommand(namespaceCmd)
}
