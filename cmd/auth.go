package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "认证管理",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录 GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		adapter := getAdapter()
		if err := adapter.Login(); err != nil {
			fmt.Fprintf(os.Stderr, "登录失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "登录成功")
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "登出",
	Run: func(cmd *cobra.Command, args []string) {
		adapter := getAdapter()
		if err := adapter.Logout(); err != nil {
			fmt.Fprintf(os.Stderr, "登出失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "已登出")
	},
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "显示当前登录状态",
	Run: func(cmd *cobra.Command, args []string) {
		adapter := getAdapter()
		status, err := adapter.GetAuthStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取状态失败: %v\n", err)
			os.Exit(1)
		}

		out := cmd.OutOrStdout()
		if !status.LoggedIn {
			fmt.Fprintln(out, "未登录")
			fmt.Fprintln(out, "\n使用 'dehub auth login' 登录")
			return
		}

		fmt.Fprintf(out, "已登录: %s\n", status.Username)
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authWhoamiCmd)
	rootCmd.AddCommand(authCmd)
}
