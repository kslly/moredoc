/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"moredoc/pkg/logger"
	"moredoc/service"

	"github.com/spf13/cobra"
)

// migrateContentCmd represents the migrateContent command
var migrateContentCmd = &cobra.Command{
	Use:   "migrateContent",
	Short: "迁移内容",
	Long:  `将文档文本内容迁移到数据库中`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.NewLogger()
		if err != nil {
			log.Print("instantiation logger error: ", err)
			return
		}
		service.MigrateContent(cfg, lg)
	},
}

func init() {
	rootCmd.AddCommand(migrateContentCmd)
}
