package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-google-cpi/action"
	"bosh-google-cpi/api"
	"bosh-google-cpi/api/dispatcher"
	"bosh-google-cpi/api/transport"
	"bosh-google-cpi/config"

	"bosh-google-cpi/google/client"
)

const mainLogTag = "main"

var (
	configFileOpt = flag.String("configFile", "", "Path to configuration file")
	input         io.Reader
	output        io.Writer
)

func main() {
	logger, fs, cmdRunner, uuidGen := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	config, err := config.NewConfigFromPath(*configFileOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config - %s", err.Error())
		os.Exit(1)
	}

	dispatcher, err := buildDispatcher(config, logger, fs, cmdRunner, uuidGen)
	if err != nil {
		logger.Error(mainLogTag, "Building Dispatcher - %s", err)
		os.Exit(1)
	}

	cli := transport.NewCLI(os.Stdin, os.Stdout, dispatcher, logger)

	if err = cli.ServeOnce(); err != nil {
		logger.Error(mainLogTag, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (api.MultiLogger, boshsys.FileSystem, boshsys.CmdRunner, boshuuid.Generator) {
	var logBuff bytes.Buffer
	multiWriter := io.MultiWriter(os.Stderr, bufio.NewWriter(&logBuff))
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, multiWriter, os.Stderr)
	multiLogger := api.MultiLogger{Logger: logger, LogBuff: &logBuff}
	fs := boshsys.NewOsFileSystem(multiLogger)

	cmdRunner := boshsys.NewExecCmdRunner(multiLogger)

	uuidGen := boshuuid.NewGenerator()

	return multiLogger, fs, cmdRunner, uuidGen
}

func buildDispatcher(
	config config.Config,
	logger api.MultiLogger,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
) (dispatcher.Dispatcher, error) {
	googleClient, err := client.NewGoogleClient(config.Google, logger)
	if err != nil {
		return nil, err
	}

	actionFactory := action.NewConcreteFactory(
		googleClient,
		uuidGen,
		config.Actions,
		logger,
	)

	caller := dispatcher.NewJSONCaller()

	return dispatcher.NewJSON(actionFactory, caller, logger), nil
}
