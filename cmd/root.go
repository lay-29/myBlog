package cmd

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"myblog/modules/setting"
	"myblog/routers"
	"myblog/services"
)

const defaultConfigPath = "custom/conf/app.ini"

func Execute(args []string) error {
	if len(args) == 0 {
		return runWeb(nil)
	}

	switch args[0] {
	case "web":
		return runWeb(args[1:])
	case "admin":
		return reservedCommand("admin")
	case "migrate":
		return reservedCommand("migrate")
	case "doctor":
		return reservedCommand("doctor")
	case "indexer":
		if len(args) > 1 && args[1] == "rebuild" {
			return reservedCommand("indexer rebuild")
		}
		return errors.New("usage: myblog indexer rebuild")
	case "-h", "--help", "help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runWeb(args []string) error {
	flags := flag.NewFlagSet("web", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)

	configPath := flags.String("c", defaultConfigPath, "config file path")
	flags.StringVar(configPath, "config", defaultConfigPath, "config file path")
	workPath := flags.String("work-path", ".", "application work path")

	if err := flags.Parse(args); err != nil {
		return err
	}

	absWorkPath, err := filepath.Abs(*workPath)
	if err != nil {
		return err
	}

	cfg, err := setting.Load(*configPath, absWorkPath)
	if err != nil {
		return err
	}

	app, err := services.NewApp(cfg)
	if err != nil {
		return err
	}
	defer app.Close()

	handler, err := routers.New(app)
	if err != nil {
		return err
	}

	addr := cfg.Server.ListenAddr()
	fmt.Printf("%s listening on http://%s\n", cfg.AppName, addr)
	return http.ListenAndServe(addr, handler)
}

func reservedCommand(name string) error {
	return fmt.Errorf("%s command is reserved in this skeleton; implement subcommand behavior on top of cmd.Execute", name)
}

func printHelp() {
	fmt.Println(`MyBlog - self-hosted knowledge base and blog

Usage:
  myblog web -c custom/conf/app.ini --work-path .
  myblog admin
  myblog migrate
  myblog doctor
  myblog indexer rebuild`)
}
