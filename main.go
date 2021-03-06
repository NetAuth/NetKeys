package main

import (
	"fmt"
	"os"
	"strings"
	"context"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/hashicorp/go-hclog"

	"github.com/netauth/netauth/pkg/netauth"
)

var (
	keyType   = pflag.String("type", "SSH", "Type of keys to print")
	entityID  = pflag.String("ID", "", "ID to look up")
	serviceID = pflag.String("service", "netkeys", "Service ID to send")
	cfgfile   = pflag.String("config", "", "Config file to use")
	verbose   = pflag.Bool("verbose", false, "Show logs")
)

func main() {
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	if *cfgfile != "" {
		viper.SetConfigFile(*cfgfile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.netauth")
		viper.AddConfigPath("/etc/netauth/")
	}
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	// Shut off all the logging
	if !*verbose {
		hclog.SetDefault(hclog.NewNullLogger())
	}
	l := hclog.L().Named("netkeys")

	c, err := netauth.New()
	if err != nil {
		l.Warn("Error during client initialization:", "error", err)
		os.Exit(1)
	}

	// Set the service ID
	c.SetServiceName(*serviceID)

	e, err := c.EntityInfo(context.Background(), *entityID)
	if err != nil {
		l.Error("Error loading entity:", "error", err)
		os.Exit(1)
	}
	if e.GetMeta().GetLocked() {
		// If locked metadata is present, then don't return
		// anything.
		os.Exit(0)
	}

	// This is only ever done for read, never write, so we feed a
	// null token
	keys, err := c.EntityKeys(context.Background(), *entityID, "READ", *keyType, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Print out the keys, no formatting, just the key data
	for _, k := range keys[strings.ToUpper(*keyType)] {
		fmt.Println(k)
	}
	os.Exit(0)
}
