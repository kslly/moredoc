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

// fixCoverCmd represents the fixCover command
var fixCoverCmd = &cobra.Command{
	Use:   "fixCover",
	Short: "修正封面大小",
	Long:  `修正已有图片的封面大小，使其符合要求，特别是PPT类的文档。`,
	Run: func(cmd *cobra.Command, args []string) {
		lg, err := logger.NewLogger()
		if err != nil {
			log.Print("instantiation logger error: ", err)
			return
		}
		service.FixCover(cfg, lg)
	},
}

func init() {
	rootCmd.AddCommand(fixCoverCmd)
}
